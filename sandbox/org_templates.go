// Gives each organization its own substrate template: a copy of the base template at the
// organization's size, running serve's runner image, so substrate holds its environments to that
// size and tells them apart, and deletes the copies no organization runs from any more.
package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrgTemplates makes and names each organization's template on substrate.
type OrgTemplates struct {
	DB     *gorm.DB
	Client *substrate.Client
	// Atespace is where the templates and the environments created from them are.
	Atespace string
	// Base is the template every organization's is a copy of, which sets the image runs execute.
	Base string
	// Size is an organization's size unless a platform admin set one of its own; a field left
	// zero is the base template's.
	Size substrate.Size
	// Image is the runner image the organizations' copies run in place of the base template's: the
	// runner of serve's own release, so each release's runs execute its runner without the base
	// being registered again. Empty runs the base template's image.
	Image string
	// Gavel is this gavel's name, which starts its organizations' template names, so gavels
	// sharing a substrate cluster each sweep only their own.
	Gavel string

	// mu keeps Sweep from deleting a template Resolve is handing a run.
	mu sync.Mutex
}

// OrgTemplate is the template an organization's runs are created from, and the size it gives them.
type OrgTemplate struct {
	Name string
	Size substrate.Size
}

// prefix starts the name of every template of this gavel's organizations.
func (o *OrgTemplates) prefix() string {
	return "gavel-" + o.Gavel + "-"
}

// MaxGavelNameLength is the longest gavel name whose organizations' template names, the prefix,
// the organization's 32 hex digits and an 8-digit hash, still fit a 63-character DNS label.
const MaxGavelNameLength = 15

// Resolve is org's template, which it makes from the base on first use. Its name names the base
// template's UID, the size and the runner image, so a new base, a new size or a new release's
// runner gives the organization a new template: a template can't change.
func (o *OrgTemplates) Resolve(ctx context.Context, org uuid.UUID) (OrgTemplate, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	base, err := o.Client.Template(ctx, o.Base)
	if err != nil {
		return OrgTemplate{}, fmt.Errorf("give organization %s its substrate template: %w", org, err)
	}
	template, err := o.wanted(ctx, org, base)
	if err != nil {
		return OrgTemplate{}, err
	}
	if _, err := o.Client.CopyTemplate(ctx, o.Base, template.Name, template.Size, o.Image); err != nil && !errors.Is(err, substrate.ErrTemplateExists) {
		return OrgTemplate{}, fmt.Errorf("give organization %s its substrate template: %w", org, err)
	}
	return template, nil
}

// DefaultSize is the size of an organization without one of its own: Size, with what it leaves
// unset from the base template.
func (o *OrgTemplates) DefaultSize(ctx context.Context) (substrate.Size, error) {
	base, err := o.Client.Template(ctx, o.Base)
	if err != nil {
		return substrate.Size{}, err
	}
	return o.defaultSize(base), nil
}

func (o *OrgTemplates) defaultSize(base substrate.Template) substrate.Size {
	size := o.Size
	if size.CPUMillicores == 0 {
		size.CPUMillicores = base.Size.CPUMillicores
	}
	if size.MemoryBytes == 0 {
		size.MemoryBytes = base.Size.MemoryBytes
	}
	return size
}

// wanted is the template org's runs should be created from, given the base template.
func (o *OrgTemplates) wanted(ctx context.Context, org uuid.UUID, base substrate.Template) (OrgTemplate, error) {
	size, err := OrgSize(ctx, o.DB, org)
	if err != nil {
		return OrgTemplate{}, err
	}
	if size == nil {
		defaults := o.defaultSize(base)
		size = &defaults
	}
	sum := sha256.Sum256([]byte(base.UID + "/" + strconv.FormatInt(size.CPUMillicores, 10) + "/" + strconv.FormatInt(size.MemoryBytes, 10) + "/" + o.Image))
	return OrgTemplate{Name: o.prefix() + hex.EncodeToString(org[:]) + "-" + hex.EncodeToString(sum[:4]), Size: *size}, nil
}

// Sweep deletes the organizations' templates that no actor runs and that are not their
// organization's current one, or whose organization gavel no longer has, and returns their names.
// A template an actor runs stays: env can't delete an environment whose template is gone. An
// organization that hasn't run in a while keeps its current template, which, without a golden
// snapshot, holds nothing.
func (o *OrgTemplates) Sweep(ctx context.Context) ([]string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	base, err := o.Client.Template(ctx, o.Base)
	if err != nil {
		return nil, err
	}
	templates, err := o.Client.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}
	actors, err := o.Client.ListActors(ctx, o.Atespace)
	if err != nil {
		return nil, err
	}
	running := map[string]bool{}
	for _, actor := range actors {
		running[actor.Template] = true
	}
	var swept []string
	var errs []error
	for _, template := range templates {
		org, ok := o.templateOrg(template.Name)
		if !ok || running[template.Name] {
			continue
		}
		wanted, err := o.wanted(ctx, org, base)
		if err != nil && !errors.Is(err, ErrOrgNotFound) {
			errs = append(errs, err)
			continue
		}
		if err == nil && wanted.Name == template.Name {
			continue
		}
		if err := o.Client.DeleteTemplate(ctx, template.Name); err != nil {
			errs = append(errs, err)
			continue
		}
		swept = append(swept, template.Name)
	}
	return swept, errors.Join(errs...)
}

// templateOrg is the organization of this gavel whose template name is, when it is one.
func (o *OrgTemplates) templateOrg(name string) (uuid.UUID, bool) {
	rest, ok := strings.CutPrefix(name, o.prefix())
	if !ok || len(rest) <= 32 || rest[32] != '-' {
		return uuid.Nil, false
	}
	org, err := uuid.Parse(rest[:32])
	return org, err == nil
}

// ErrOrgNotFound reports an organization gavel doesn't have.
var ErrOrgNotFound = errors.New("organization not found")

// ErrInvalidOrgSize refuses a size without both CPU and memory, which substrate needs.
var ErrInvalidOrgSize = errors.New("an organization's sandbox size needs both CPU and memory")

// OrgSize is the size a platform admin set for org, or nil.
func OrgSize(ctx context.Context, db *gorm.DB, org uuid.UUID) (*substrate.Size, error) {
	var row struct {
		SandboxCPUMillicores *int64 `gorm:"column:sandbox_cpu_millicores"`
		SandboxMemoryBytes   *int64 `gorm:"column:sandbox_memory_bytes"`
	}
	result := db.WithContext(ctx).Raw(`SELECT sandbox_cpu_millicores, sandbox_memory_bytes FROM orgs WHERE id = ?`, org).Scan(&row)
	if result.Error != nil {
		return nil, fmt.Errorf("read the sandbox size of organization %s: %w", org, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("read the sandbox size of organization %s: %w", org, ErrOrgNotFound)
	}
	if row.SandboxCPUMillicores == nil || row.SandboxMemoryBytes == nil {
		return nil, nil
	}
	return &substrate.Size{CPUMillicores: *row.SandboxCPUMillicores, MemoryBytes: *row.SandboxMemoryBytes}, nil
}

// SetOrgSize sets the size of org's sandbox environments on substrate; nil returns it to serve's
// default. Its runs from then on get a template of that size.
func SetOrgSize(ctx context.Context, db *gorm.DB, org uuid.UUID, size *substrate.Size) error {
	var cpu, memory *int64
	if size != nil {
		if size.CPUMillicores <= 0 || size.MemoryBytes <= 0 {
			return fmt.Errorf("set the sandbox size of organization %s to %d millicores and %d bytes: %w",
				org, size.CPUMillicores, size.MemoryBytes, ErrInvalidOrgSize)
		}
		cpu, memory = &size.CPUMillicores, &size.MemoryBytes
	}
	result := db.WithContext(ctx).Exec(`UPDATE orgs SET sandbox_cpu_millicores = ?, sandbox_memory_bytes = ?, updated_at = now() WHERE id = ?`, cpu, memory, org)
	if result.Error != nil {
		return fmt.Errorf("set the sandbox size of organization %s: %w", org, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("set the sandbox size of organization %s: %w", org, ErrOrgNotFound)
	}
	return nil
}
