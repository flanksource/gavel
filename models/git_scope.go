package models

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

type Scopes []ScopeType

func (s Scopes) Contains(scope ScopeType) bool {
	for _, sc := range s {
		if sc == scope {
			return true
		}
	}
	return false
}

func (s Scopes) ToString() []string {
	var result []string
	for _, scope := range s {
		result = append(result, string(scope))
	}
	return result
}

type ScopeType string

const (
	ScopeTypeDocs    ScopeType = "docs"
	ScopeTypeGeneral ScopeType = "general"
	// Continuous Integration, build pipelines, image builds
	ScopeTypeCI            ScopeType = "ci"
	ScopeTypeMicroservices ScopeType = "microservices"
	// Application monitoring and observability e.g. new Prometheus metrics, tracing, logging
	ScopeTypeObservability ScopeType = "observability"
	ScopeTypeNetworking    ScopeType = "networking"
	ScopeTypeSecurity      ScopeType = "security"
	ScopeTypeDatabase      ScopeType = "database"
	// Infrastructure instances e.g. Terraform module usage, helm chart values,
	ScopeTypeInfrastructure ScopeType = "infrastructure"
	// Infrastructure code definitions, e.g., Terraform, Ansible, CloudFormation, Helm chart definitions
	ScopeTypeIaC        ScopeType = "iac"
	ScopeTypeApp        ScopeType = "app"
	ScopeTypeDeployment ScopeType = "deployment"
	ScopeTypeCD                   = ScopeTypeDeployment
	// Scaling operations, e.g., autoscaling, load balancing
	ScopeTypeScaling          ScopeType = "scaling"
	ScopeTypeTest             ScopeType = "test"
	ScopeTypeReliability      ScopeType = "reliability"
	ScopeTypePerformance      ScopeType = "performance"
	ScopeTypeCostOptimization ScopeType = "cost_optimization"
	ScopeTypeSecrets          ScopeType = "secrets"
	ScopeTypeConfig           ScopeType = "config"
	ScopeTypeDependency       ScopeType = "dependency"
	ScopeTypeML               ScopeType = "ml"
	ScopeTypeOther            ScopeType = "other"
	ScopeTypeUnknown          ScopeType = ""
)

func (s ScopeType) Pretty() api.Text {
	t := clicky.Text("")

	switch s {
	case ScopeTypeTest:
		return t.Add(icons.Test)
	case ScopeTypeDocs:
		return t.Add(icons.Docs)
	case ScopeTypeDeployment:
		return t.Add(icons.Launch)
	case ScopeTypeCI:
		return t.Add(icons.CI)
	case ScopeTypeSecurity:
		return t.Add(icons.Lock)
	case ScopeTypeDatabase:
		return t.Add(icons.DB)
	case ScopeTypeIaC:
		return t.Add(icons.Infrastructure)
	case ScopeTypeNetworking:
		return t.Add(icons.Network)
	case ScopeTypeObservability:
		return t.Add(icons.Monitor)
	case ScopeTypeInfrastructure:
		return t.Add(icons.Infrastructure)
	case ScopeTypeScaling:
		return t.Add(icons.Scaling)
	case ScopeTypeReliability:
		return t.Add(icons.Reliability)
	case ScopeTypePerformance:
		return t.Add(icons.Performance)
	case ScopeTypeCostOptimization:
		return t.Add(icons.Cost)
	case ScopeTypeSecrets:
		return t.Add(icons.Key)
	case ScopeTypeConfig:
		return t.Add(icons.Config)
	case ScopeTypeDependency:
		return t.Add(icons.Dependency)
	case ScopeTypeML:
		return t.Add(icons.AI)
	case ScopeTypeOther:
		return t.Add(icons.Package)
	}

	return t.Append(string(s))
}

type Technology []ScopeTechnology

func (t Technology) ToString() []string {
	var result []string
	for _, tech := range t {
		result = append(result, string(tech))
	}
	return result
}

type ScopeTechnology string

const (
	Kubernetes    ScopeTechnology = "kubernetes"
	Bazel         ScopeTechnology = "bazel"
	Docker        ScopeTechnology = "docker"
	Terraform     ScopeTechnology = "terraform"
	Markdown      ScopeTechnology = "markdown"
	Prometheus    ScopeTechnology = "prometheus"
	Grafana       ScopeTechnology = "grafana"
	Jenkins       ScopeTechnology = "jenkins"
	Ansible       ScopeTechnology = "ansible"
	Helm          ScopeTechnology = "helm"
	GitOps        ScopeTechnology = "gitops"
	AWS           ScopeTechnology = "aws"
	GCP           ScopeTechnology = "gcp"
	Azure         ScopeTechnology = "azure"
	Linux         ScopeTechnology = "linux"
	Openshift     ScopeTechnology = "openshift"
	MongoDB       ScopeTechnology = "mongodb"
	PostgreSQL    ScopeTechnology = "postgresql"
	MySQL         ScopeTechnology = "mysql"
	Redis         ScopeTechnology = "redis"
	Nginx         ScopeTechnology = "nginx"
	Clickhouse    ScopeTechnology = "clickhouse"
	Kafka         ScopeTechnology = "kafka"
	Cassandra     ScopeTechnology = "cassandra"
	Gitlab        ScopeTechnology = "gitlab"
	ArgoCD        ScopeTechnology = "argocd"
	FluxCD        ScopeTechnology = "fluxcd"
	OpenTelemetry ScopeTechnology = "opentelemetry"
	GitHubActions ScopeTechnology = "github_actions"
	Python        ScopeTechnology = "python"
	Java          ScopeTechnology = "java"
	Ruby          ScopeTechnology = "ruby"
	Rust          ScopeTechnology = "rust"
	PHP           ScopeTechnology = "php"
	NodeJS        ScopeTechnology = "nodejs"
	Go            ScopeTechnology = "go"
	Shell         ScopeTechnology = "shell"
	Powershell    ScopeTechnology = "powershell"
	React         ScopeTechnology = "react"
	Bash          ScopeTechnology = "bash"
	SQL           ScopeTechnology = "sql"
)

type CommitType string

const (
	CommitTypeFeat       CommitType = "feat"
	CommitTypeFix        CommitType = "fix"
	CommitTypeChore      CommitType = "chore"
	CommitTypeDocs       CommitType = "docs"
	CommitTypeStyle      CommitType = "style"
	CommitTypeRefactor   CommitType = "refactor"
	CommitTypePerf       CommitType = "perf"
	CommitTypeTest       CommitType = "test"
	CommitTypeCi         CommitType = "ci"
	CommitTypeBuild      CommitType = "build"
	CommitTypeRevert     CommitType = "revert"
	CommitTypeConfig     CommitType = "config"
	CommitTypeOther      CommitType = "other"
	CommitTypeSecurity   CommitType = "security"
	CommitTypeDependency CommitType = "dependency"
	CommitTypeUnknown    CommitType = ""
)

func (ct CommitType) Pretty() api.Text {
	t := clicky.Text("")

	switch ct {
	case CommitTypeFeat:
		return t.Add(icons.Feat)
	case CommitTypeFix:
		return t.Add(icons.Fix)
	case CommitTypeChore:
		return t.Add(icons.Chore)
	case CommitTypeDocs:
		return t.Add(icons.Docs)
	case CommitTypeStyle:
		return t.Add(icons.Style)
	case CommitTypeRefactor:
		return t.Add(icons.Refactor)
	case CommitTypePerf:
		return t.Add(icons.Performance)
	case CommitTypeTest:
		return t.Add(icons.Test)
	case CommitTypeCi:
		return t.Add(icons.CI)
	case CommitTypeBuild:
		return t.Add(icons.CI)
	case CommitTypeRevert:
		return t.Add(icons.Undo)
	case CommitTypeConfig:
		return t.Add(icons.Config)
	case CommitTypeSecurity:
		return t.Add(icons.Lock)
	case CommitTypeDependency:
		return t.Add(icons.Dependency)
	}

	return t.Append(string(ct))
}
