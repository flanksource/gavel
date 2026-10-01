package archtest

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
)

// Gavel never writes a Captain table. Captain owns every write to captain_*
// and exposes domain APIs and hooks for the rest; Gavel only reads through it
// and writes its own todo_* rows (for instance from a Captain Link callback).
// This suite fails on any non-test Gavel code or SQL that writes one anyway.

const captainDatabasePackage = "github.com/flanksource/captain/pkg/database"

// mutatingCaptainMethod is the vocabulary Captain's *database.DB uses for a
// method that writes.
var mutatingCaptainMethod = regexp.MustCompile(`^(Create|Update|Upsert|Append|Set|Approve|Resolve|Cancel|Put|Register|Delete|Mark|Expire|Hold|Release|Lock|Insert|Touch|Finish|Link|Select|Attach|Save)`)

// allowedCaptainMethods are the host hooks Captain provides for exactly this
// boundary. Transaction only scopes calls to Captain's own APIs; a write made
// through the scoped handle is still caught by the method check.
var allowedCaptainMethods = map[string]bool{
	"RegisterDeleteGuard": true,
	"ListenRowChanges":    true,
	"Transaction":         true,
}

// captainWrite matches SQL that writes a Captain table or installs DDL on one.
// The table must follow the keyword directly, so a Gavel table or constraint
// merely named after Captain (todo_issue_prompt_runs_captain_…) never matches.
var captainWrite = regexp.MustCompile(`(?is)\b(` +
	`INSERT\s+INTO|` +
	`UPDATE|` +
	`DELETE\s+FROM|` +
	`TRUNCATE(\s+TABLE)?|` +
	`ALTER\s+TABLE(\s+IF\s+EXISTS)?|` +
	`REFERENCES|` +
	`CREATE\s+(OR\s+REPLACE\s+)?(CONSTRAINT\s+)?TRIGGER\s+[^;]*?\bON|` +
	`CREATE\s+(UNIQUE\s+)?INDEX\s+[^;]*?\bON` +
	`)\s+(ONLY\s+)?("?public"?\.)?"?captain_\w+`)

var sqlLineComment = regexp.MustCompile(`--[^\n]*`)

// ContinueOnFailure: a package that fails to load fails the suite on its own
// spec, and every offender in the packages that did load is still listed.
var _ = Describe("Gavel's writes to Captain tables", Ordered, ContinueOnFailure, func() {
	var root string
	var gavel []*packages.Package

	BeforeAll(func() {
		var err error
		root, err = filepath.Abs(filepath.Join("..", ".."))
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Join(root, "go.mod")).To(BeARegularFile(), "the guard must run from the gavel module root")
		gavel, err = packages.Load(&packages.Config{
			Dir:   root,
			Tests: false,
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
				packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		}, "./...")
		Expect(err).NotTo(HaveOccurred())
		Expect(gavel).NotTo(BeEmpty())
	})

	It("loads every non-test package, so no offender can hide in one that does not compile", func() {
		Expect(loadErrors(root, gavel)).To(BeEmpty())
	})

	It("calls no mutating method of Captain's *database.DB outside the host hooks", func() {
		Expect(mutatingCaptainCalls(root, gavel)).To(BeEmpty())
	})

	It("embeds no SQL that writes a Captain table in Go code", func() {
		Expect(captainWritesInGoStrings(root, gavel)).To(BeEmpty())
	})

	It("ships no SQL file that writes a Captain table or installs DDL on one", func() {
		offenders, err := captainWritesInSQLFiles(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(offenders).To(BeEmpty())
	})
})

var _ = DescribeTable("recognising SQL that writes a Captain table",
	func(sql string, want []string) {
		Expect(captainWrites(sql)).To(Equal(want))
	},
	Entry("an insert", "INSERT INTO captain_sessions (id) VALUES ($1)",
		[]string{"INSERT INTO captain_sessions"}),
	Entry("a schema-qualified update", "UPDATE public.captain_prompt_runs SET state = 'failed'",
		[]string{"UPDATE public.captain_prompt_runs"}),
	Entry("a delete split across lines", "DELETE\n  FROM captain_turn_requests WHERE id = $1",
		[]string{"DELETE FROM captain_turn_requests"}),
	Entry("a trigger on a Captain table", "CREATE OR REPLACE TRIGGER t\nAFTER INSERT OR UPDATE OF state\nON public.captain_prompt_runs FOR EACH ROW",
		[]string{"CREATE OR REPLACE TRIGGER t AFTER INSERT OR UPDATE OF state ON public.captain_prompt_runs"}),
	Entry("a foreign key into a Captain table", "ADD CONSTRAINT fk FOREIGN KEY (plan_id) REFERENCES public.captain_plans (id)",
		[]string{"REFERENCES public.captain_plans"}),
	Entry("altering a Captain table", "ALTER TABLE IF EXISTS captain_plans ADD COLUMN x int",
		[]string{"ALTER TABLE IF EXISTS captain_plans"}),
	Entry("a Gavel table whose constraint is named after Captain",
		"ALTER TABLE public.todo_issue_prompt_runs DROP CONSTRAINT IF EXISTS todo_issue_prompt_runs_captain_prompt_run_fkey", nil),
	Entry("a Gavel update that reads a Captain table",
		"UPDATE public.todo_issues AS issue SET status = 'open' FROM public.captain_plans AS plan WHERE issue.selected_plan_id = plan.id", nil),
	Entry("a row lock on a Captain table", "SELECT id FROM captain_plans WHERE id = $1 FOR UPDATE", nil),
	Entry("prose in a comment", "-- UPDATE captain_prompt_runs used to be ours\nSELECT 1", nil),
)

func loadErrors(root string, pkgs []*packages.Package) []string {
	var errs []string
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		if !strings.HasPrefix(pkg.PkgPath, "github.com/flanksource/gavel") {
			return
		}
		for _, err := range pkg.Errors {
			errs = append(errs, relative(root, err.Error()))
		}
	})
	sort.Strings(errs)
	return errs
}

func mutatingCaptainCalls(root string, pkgs []*packages.Package) []string {
	var offenders []string
	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}
		for expr, selection := range pkg.TypesInfo.Selections {
			method, ok := selection.Obj().(*types.Func)
			if !ok || !isCaptainDBMethod(method) {
				continue
			}
			name := method.Name()
			if allowedCaptainMethods[name] || !mutatingCaptainMethod.MatchString(name) {
				continue
			}
			offenders = append(offenders, fmt.Sprintf("%s: (*database.DB).%s",
				position(root, pkg.Fset, expr.Sel.Pos()), name))
		}
	}
	sort.Strings(offenders)
	return offenders
}

// isCaptainDBMethod reports a method declared on Captain's DB, however it is
// reached: directly, through a Transaction-scoped handle, or promoted through
// a Gavel type embedding it.
func isCaptainDBMethod(method *types.Func) bool {
	signature, ok := method.Type().(*types.Signature)
	if !ok || signature.Recv() == nil {
		return false
	}
	receiver := signature.Recv().Type()
	if pointer, ok := receiver.(*types.Pointer); ok {
		receiver = pointer.Elem()
	}
	named, ok := receiver.(*types.Named)
	return ok && named.Obj().Pkg() != nil &&
		named.Obj().Pkg().Path() == captainDatabasePackage && named.Obj().Name() == "DB"
}

func captainWritesInGoStrings(root string, pkgs []*packages.Package) []string {
	var offenders []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				literal, ok := node.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return true
				}
				value, err := strconv.Unquote(literal.Value)
				if err != nil {
					offenders = append(offenders, fmt.Sprintf("%s: unreadable string literal: %v",
						position(root, pkg.Fset, literal.Pos()), err))
					return true
				}
				for _, statement := range captainWrites(value) {
					offenders = append(offenders, fmt.Sprintf("%s: %s", position(root, pkg.Fset, literal.Pos()), statement))
				}
				return true
			})
		}
	}
	sort.Strings(offenders)
	return offenders
}

func captainWritesInSQLFiles(root string) ([]string, error) {
	var offenders []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && skippedDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, statement := range captainWrites(string(content)) {
			offenders = append(offenders, fmt.Sprintf("%s: %s", relative(root, path), statement))
		}
		return nil
	})
	sort.Strings(offenders)
	return offenders, err
}

// skippedDir excludes dependency trees, build output, scratch space and test
// fixtures: none of it is SQL Gavel ships or runs against its database.
func skippedDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules" || name == "testdata" ||
		name == "dist" || name == "vendor" || name == "hack"
}

func captainWrites(sql string) []string {
	matches := captainWrite.FindAllString(sqlLineComment.ReplaceAllString(sql, ""), -1)
	for i, match := range matches {
		matches[i] = strings.Join(strings.Fields(match), " ")
	}
	return matches
}

func position(root string, fset *token.FileSet, pos token.Pos) string {
	p := fset.Position(pos)
	return fmt.Sprintf("%s:%d", relative(root, p.Filename), p.Line)
}

func relative(root, path string) string {
	return strings.ReplaceAll(path, root+string(filepath.Separator), "")
}
