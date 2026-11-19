// Gives each organization its own substrate template: a copy of the base template at the
// organization's size, so substrate holds its environments to that size and tells them apart.
package sandbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

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
}

// OrgTemplate is the template an organization's runs are created from, and the size it gives them.
type OrgTemplate struct {
	Name string
	Size substrate.Size
}

// orgTemplatePrefix starts the name of every organization's template.
const orgTemplatePrefix = "gavel-org-"

// Resolve is org's template, which it makes from the base on first use. Its name names the base
// template's UID and the size, so a new base, registered for the runner's next image, or a new
// size gives the organization a new template: a template can't change.
func (o *OrgTemplates) Resolve(ctx context.Context, org uuid.UUID) (OrgTemplate, error) {
	base, err := o.Client.Template(ctx, o.Base)
	if err != nil {
		return OrgTemplate{}, fmt.Errorf("give organization %s its substrate template: %w", org, err)
	}
	template, err := o.wanted(ctx, org, base)
	if err != nil {
		return OrgTemplate{}, err
	}
	if _, err := o.Client.CopyTemplate(ctx, o.Base, template.Name, template.Size); err != nil && !errors.Is(err, substrate.ErrTemplateExists) {
		return OrgTemplate{}, fmt.Errorf("give organization %s its substrate template: %w", org, err)
	}
	return template, nil
}

// wanted is the template org's runs should be created from, given the base template.
func (o *OrgTemplates) wanted(ctx context.Context, org uuid.UUID, base substrate.Template) (OrgTemplate, error) {
	size, err := orgSize(ctx, o.DB, org)
	if err != nil {
		return OrgTemplate{}, err
	}
	if size == nil {
		size = &substrate.Size{CPUMillicores: o.Size.CPUMillicores, MemoryBytes: o.Size.MemoryBytes}
		if size.CPUMillicores == 0 {
			size.CPUMillicores = base.Size.CPUMillicores
		}
		if size.MemoryBytes == 0 {
			size.MemoryBytes = base.Size.MemoryBytes
		}
	}
	sum := sha256.Sum256([]byte(base.UID + "/" + strconv.FormatInt(size.CPUMillicores, 10) + "/" + strconv.FormatInt(size.MemoryBytes, 10)))
	return OrgTemplate{Name: orgTemplatePrefix + org.String() + "-" + hex.EncodeToString(sum[:4]), Size: *size}, nil
}

// ErrOrgNotFound reports an organization gavel doesn't have.
var ErrOrgNotFound = errors.New("organization not found")

// orgSize is the size a platform admin set for org, or nil.
func orgSize(ctx context.Context, db *gorm.DB, org uuid.UUID) (*substrate.Size, error) {
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
			return fmt.Errorf("set the sandbox size of organization %s to %d millicores and %d bytes: an organization's sandbox size needs both CPU and memory",
				org, size.CPUMillicores, size.MemoryBytes)
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
