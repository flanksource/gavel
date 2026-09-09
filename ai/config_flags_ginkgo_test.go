package ai

import (
	"github.com/flanksource/captain/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

var _ = Describe("explicit AI flag request", func() {
	It("leaves advertised defaults unset and retains explicitly supplied zero values", func() {
		config := DefaultConfig()
		flags := pflag.NewFlagSet("runtime", pflag.ContinueOnError)
		BindFlags(flags, &config)
		Expect(FlagSpec(config, flags)).To(Equal(api.Spec{}))
		Expect(flags.Parse([]string{"--ai-model=api:gpt-4o", "--ai-max-tokens=0", "--ai-no-cache=false"})).To(Succeed())
		request := FlagSpec(config, flags)
		Expect(request.Model.Name).To(Equal("gpt-4o"))
		Expect(request.Model.Mode).To(Equal(api.ModeAPI))
		Expect(request.Explicit).To(HaveKey("/budget/maxTokens"))
		Expect(request.Explicit).To(HaveKey("/noCache"))
	})
})
