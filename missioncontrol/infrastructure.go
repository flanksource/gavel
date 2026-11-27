// Reads gavel's infrastructure from Mission Control's catalog: the config items, health
// checks and config changes its search selects, for the platform dashboard.
package missioncontrol

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/mission-control-sdk/go/clientapi"
)

// infrastructureLimit bounds how many resources of each kind one read returns.
const infrastructureLimit = 200

// changesDays is how many days back the changes read reach.
const changesDays = 7

// Infrastructure is what Mission Control knows about gavel's infrastructure. Each list is read on
// its own, so one Mission Control can't read leaves the others, with the reason in its Error field.
type Infrastructure struct {
	MissionControl string                       `json:"missionControl"`
	Configs        []clientapi.SelectedResource `json:"configs"`
	ConfigsError   string                       `json:"configsError,omitempty"`
	Checks         []clientapi.SelectedResource `json:"checks"`
	ChecksError    string                       `json:"checksError,omitempty"`
	Changes        []clientapi.ConfigChangeRow  `json:"changes"`
	ChangesDays    int                          `json:"changesDays"`
	ChangesError   string                       `json:"changesError,omitempty"`
}

// Infrastructure reads the config items and health checks Search selects, and the changes of
// those config items: Mission Control can't search changes by most of the fields that select
// config items, such as namespace. It fails only when neither config items nor checks are read.
func (s *Source) Infrastructure(ctx context.Context) (Infrastructure, error) {
	selector := []clientapi.ResourceSelector{{Search: s.Search}}
	configs, configsErr := s.Client.SearchCatalog(ctx, clientapi.SearchResourcesRequest{Limit: infrastructureLimit, Timestamps: true, Configs: selector})
	checks, checksErr := s.Client.SearchCatalog(ctx, clientapi.SearchResourcesRequest{Limit: infrastructureLimit, Timestamps: true, Checks: selector})
	if configsErr != nil && checksErr != nil {
		err := configsErr
		if checksErr.Error() != configsErr.Error() {
			err = fmt.Errorf("config items: %w; health checks: %w", configsErr, checksErr)
		}
		return Infrastructure{}, s.explain("gavel's infrastructure", err)
	}

	found := Infrastructure{
		MissionControl: s.URL,
		Configs:        []clientapi.SelectedResource{},
		Checks:         []clientapi.SelectedResource{},
		Changes:        []clientapi.ConfigChangeRow{},
		ChangesDays:    changesDays,
	}
	if configsErr != nil {
		found.ConfigsError = s.explain("gavel's config items", configsErr).Error()
		found.ChangesError = "gavel's changes are those of its config items, which couldn't be read"
	} else {
		found.Configs = nonNil(configs.Configs)
		changes, err := s.changesOf(ctx, found.Configs)
		if err != nil {
			found.ChangesError = err.Error()
		} else {
			found.Changes = changes
		}
	}
	if checksErr != nil {
		found.ChecksError = s.explain("gavel's health checks", checksErr).Error()
	} else {
		found.Checks = nonNil(checks.Checks)
	}
	return found, nil
}

// changesOf reads the newest changes of configs themselves, not of their children, within the last
// changesDays days and at most infrastructureLimit of them.
func (s *Source) changesOf(ctx context.Context, configs []clientapi.SelectedResource) ([]clientapi.ConfigChangeRow, error) {
	if len(configs) == 0 {
		return []clientapi.ConfigChangeRow{}, nil
	}
	ids := make([]string, len(configs))
	for i, config := range configs {
		ids[i] = config.ID
	}
	found, err := s.Client.SearchCatalogChanges(ctx, clientapi.CatalogChangesSearchRequest{BaseCatalogSearch: clientapi.BaseCatalogSearch{
		CatalogID: strings.Join(ids, ","), From: fmt.Sprintf("now-%dd", changesDays), PageSize: infrastructureLimit,
		SortBy: "-created_at", Recursive: clientapi.CatalogChangeRecursiveNone,
	}})
	if err != nil {
		return nil, s.explain("the changes of gavel's config items", err)
	}
	if found.Changes == nil {
		return []clientapi.ConfigChangeRow{}, nil
	}
	return found.Changes, nil
}

func nonNil(resources []clientapi.SelectedResource) []clientapi.SelectedResource {
	if resources == nil {
		return []clientapi.SelectedResource{}
	}
	return resources
}
