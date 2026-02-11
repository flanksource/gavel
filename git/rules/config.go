package rules

import (
	"fmt"

	"github.com/flanksource/gavel/models"
	"github.com/goccy/go-yaml"
)

// SeverityConfig represents the severity rules configuration
type SeverityConfig struct {
	Default models.Severity            `yaml:"default"`
	Rules   map[string]models.Severity `yaml:"rules"` // CEL expression -> Severity
}

// DefaultSeverityConfig returns a sensible default configuration
func DefaultSeverityConfig() *SeverityConfig {
	return &SeverityConfig{
		Default: models.Medium,
		Rules: map[string]models.Severity{
			// Deletion is always critical
			`change.type == "deleted"`: models.Critical,

			// Kubernetes security-sensitive resources
			`kubernetes.kind == "Secret"`:             models.Critical,
			`kubernetes.kind == "ServiceAccount"`:     models.High,
			`kubernetes.kind == "Role"`:               models.High,
			`kubernetes.kind == "ClusterRole"`:        models.High,
			`kubernetes.kind == "RoleBinding"`:        models.High,
			`kubernetes.kind == "ClusterRoleBinding"`: models.High,

			// Network resources
			`kubernetes.kind == "Service"`:       models.High,
			`kubernetes.kind == "Ingress"`:       models.High,
			`kubernetes.kind == "NetworkPolicy"`: models.High,

			// Storage resources
			`kubernetes.kind == "PersistentVolume"`:      models.Medium,
			`kubernetes.kind == "PersistentVolumeClaim"`: models.Medium,

			// Version changes
			`kubernetes.version_downgrade != ""`:    models.High,
			`kubernetes.version_upgrade == "major"`: models.High,
			`kubernetes.has_sha_change`:             models.High,

			// Impact metrics
			`commit.line_changes > 500`:  models.Critical,
			`commit.line_changes > 100`:  models.High,
			`commit.file_count > 20`:     models.Critical,
			`commit.file_count > 10`:     models.High,
			`commit.resource_count > 25`: models.Critical,
			`commit.resource_count > 15`: models.High,
			`change.field_count > 20`:    models.High,

			// Replica scaling
			`kubernetes.replica_delta > 10`: models.High,
			`kubernetes.replica_delta < -5`: models.High,

			// Environment and resource changes
			`kubernetes.has_env_change`:      models.Medium,
			`kubernetes.has_resource_change`: models.Medium,

			// File patterns
			`file.extension == ".env"`:                    models.High,
			`file.is_config && change.type == "modified"`: models.Medium,
		},
	}
}

// LoadSeverityConfig loads severity configuration from YAML bytes
func LoadSeverityConfig(data []byte) (*SeverityConfig, error) {
	config := &SeverityConfig{
		Default: models.Medium,
		Rules:   make(map[string]models.Severity),
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse severity config: %w", err)
	}

	// Validate severities
	for expr, severity := range config.Rules {
		if severity.Value() == 0 {
			return nil, fmt.Errorf("invalid severity '%s' for rule: %s", severity, expr)
		}
	}

	if config.Default.Value() == 0 {
		config.Default = models.Medium
	}

	return config, nil
}

// Merge combines two configs, with overrides taking precedence
func (c *SeverityConfig) Merge(overrides *SeverityConfig) *SeverityConfig {
	if overrides == nil {
		return c
	}

	merged := &SeverityConfig{
		Default: c.Default,
		Rules:   make(map[string]models.Severity),
	}

	// Copy base rules
	for expr, sev := range c.Rules {
		merged.Rules[expr] = sev
	}

	// Apply overrides
	for expr, sev := range overrides.Rules {
		merged.Rules[expr] = sev
	}

	// Override default if specified
	if overrides.Default.Value() > 0 {
		merged.Default = overrides.Default
	}

	return merged
}
