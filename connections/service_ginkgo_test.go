package connections

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/flanksource/clicky/text"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConnections(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Connections secrets") }

var _ = Describe("sensitive keeper configuration", func() {
	It("reports KEK readiness and configuration errors without returning key material", func() {
		connection := text.NewSensitiveString("local_kms://" + base64.URLEncoding.EncodeToString(make([]byte, 32)))
		Expect(Status(context.Background(), connection).State).To(Equal("ready"))
		status := Status(context.Background(), text.NewSensitiveString("local_kms://invalid-key-canary"))
		Expect(status.State).To(Equal("unavailable"))
		Expect(status.Detail).NotTo(ContainSubstring("invalid-key-canary"))
	})
	It("redacts keeper options in diagnostic formatting and JSON", func() {
		const canary = "local_kms://keeper-value-canary"
		options := Options{KeeperConnection: text.NewSensitiveString(canary)}
		for _, format := range []string{"%v", "%+v", "%#v"} {
			Expect(fmt.Sprintf(format, options)).NotTo(ContainSubstring(canary))
		}
		encoded, err := json.Marshal(options)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).NotTo(ContainSubstring(canary))
	})
	It("keeps invalid inline key material out of errors", func() {
		const canary = "invalid-keeper-value-canary"
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		_, _, err := Keeper(context.Background(), text.NewSensitiveString("local_kms://"+canary), false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).NotTo(ContainSubstring(canary))
	})
})
