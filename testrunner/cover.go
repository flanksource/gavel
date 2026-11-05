package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// CoverTest reruns one Go test or Ginkgo spec alone with -coverprofile=profile, built with the run's
// tags and go test runtime, so the profile holds only what that test executed. Its outcome is not
// read; a rerun that writes no profile is an error.
func CoverTest(opts RunOptions, test parsers.Test, profile string) error {
	if test.WorkDir != "" {
		opts.WorkDir = test.WorkDir
	}
	o := &TestOrchestrator{RunOptions: opts}
	target := test.PackagePath
	if target == "" {
		target = "."
	}
	plan := goBatchPlan{targets: []string{target}, framework: test.Framework, coverProfile: profile}
	switch test.Framework {
	case parsers.GoTest:
		var segments []string
		for _, segment := range strings.Split(test.Name, "/") {
			segments = append(segments, "^"+regexp.QuoteMeta(segment)+"$")
		}
		plan.args = []string{"-run=" + strings.Join(segments, "/")}
	case parsers.Ginkgo:
		bootstrap, err := ginkgoBootstrap(filepath.Join(o.WorkDir, target))
		if err != nil {
			return err
		}
		plan.args = []string{"-run=^" + regexp.QuoteMeta(bootstrap) + "$", "-ginkgo.focus=" + ginkgoFocus(test)}
	default:
		return fmt.Errorf("per-test coverage needs a go test or ginkgo test, not %s", test.Framework)
	}
	if len(o.Tags) > 0 {
		plan.args = append(plan.args, "-tags="+strings.Join(o.Tags, ","))
	}
	process, err := o.buildGoBatchCommand(plan)
	if err != nil {
		return err
	}
	if o.Context != nil {
		process.WithContext(o.Context)
	}
	result := process.Run().Result()
	if _, err := os.Stat(profile); err != nil {
		output := ""
		if result != nil {
			output = result.Output()
		}
		return fmt.Errorf("%s %s wrote no coverage profile: %w\n%s", process.Cmd, strings.Join(process.Args, " "), err, output)
	}
	return nil
}

// ginkgoBootstrap is the name of the Go test function in dir that calls RunSpecs.
func ginkgoBootstrap(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	locations, err := parsers.BuildTestLocationMap(dir)
	if err != nil {
		return "", err
	}
	for name, location := range locations {
		if location.IsGinkgoBootstrap && filepath.Dir(location.File) == dir {
			return name, nil
		}
	}
	return "", fmt.Errorf("no Go test calling RunSpecs in %s", dir)
}

// ginkgoFocus matches exactly the spec's text, which Ginkgo's --focus compares against the suite
// description, its containers' texts and its own text joined by spaces.
func ginkgoFocus(test parsers.Test) string {
	text := strings.Join(append(append([]string(nil), test.Suite...), test.Name), " ")
	if context, ok := test.Context.(parsers.GinkgoContext); ok && context.SuiteDescription != "" {
		return "^" + regexp.QuoteMeta(context.SuiteDescription+" "+text) + "$"
	}
	return regexp.QuoteMeta(" "+text) + "$"
}
