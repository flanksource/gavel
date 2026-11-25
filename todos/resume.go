package todos

import (
	"fmt"

	captainapi "github.com/flanksource/captain/pkg/api"
)

// SameProviderFamily reports whether two recorded provider names — a prior
// run's resolved provider and the provider a new run resolved — name the same
// family. A name no family claims is an error: a resume decided on it would be
// a guess.
func SameProviderFamily(prior, requested string) (bool, error) {
	priorFamily, err := providerFamily("prior run", prior)
	if err != nil {
		return false, err
	}
	requestedFamily, err := providerFamily("requested run", requested)
	if err != nil {
		return false, err
	}
	return priorFamily.Name == requestedFamily.Name, nil
}

func providerFamily(role, name string) (*captainapi.ModelProvider, error) {
	family, ok := captainapi.ProviderByName(name)
	if !ok {
		return nil, fmt.Errorf("%s names no known provider %q (known providers: %s)", role, name, captainapi.ProviderList())
	}
	return family, nil
}
