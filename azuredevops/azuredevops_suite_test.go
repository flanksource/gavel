package azuredevops

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAzureDevOps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Azure DevOps")
}
