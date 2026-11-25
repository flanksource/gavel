package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/flanksource/gavel/comments"
)

func (c *Client) commentIteration(ctx context.Context, number int, selector comments.CodeSelector) (map[string]any, error) {
	path := c.repositoryPath() + "/pullrequests/" + strconv.Itoa(number) + "/iterations"
	iterations, err := collection[struct {
		ID     int `json:"id"`
		Source struct {
			ID string `json:"commitId"`
		} `json:"sourceRefCommit"`
		Common struct {
			ID string `json:"commitId"`
		} `json:"commonRefCommit"`
	}](ctx, c, path, nil)
	if err != nil {
		return nil, err
	}
	iteration := 0
	for _, i := range iterations {
		if i.Source.ID == selector.HeadCommit && i.Common.ID == selector.BaseCommit && i.ID > iteration {
			iteration = i.ID
		}
	}
	if iteration == 0 {
		return nil, fmt.Errorf("%w: captured diff does not match a native PR iteration", comments.ErrConflict)
	}
	skip := 0
	for {
		query := url.Values{"$compareTo": {"0"}, "$top": {"2000"}, "$skip": {strconv.Itoa(skip)}}
		data, _, err := c.request(ctx, http.MethodGet, path+"/"+strconv.Itoa(iteration)+"/changes", query, nil)
		if err != nil {
			return nil, err
		}
		var changes struct {
			Entries []struct {
				ID   int `json:"changeTrackingId"`
				Item struct {
					Path string `json:"path"`
				} `json:"item"`
				OriginalPath string `json:"originalPath"`
			} `json:"changeEntries"`
			NextSkip int `json:"nextSkip"`
			NextTop  int `json:"nextTop"`
		}
		if err = json.Unmarshal(data, &changes); err != nil {
			return nil, err
		}
		for _, entry := range changes.Entries {
			path := entry.Item.Path
			if selector.Side == "old" && entry.OriginalPath != "" {
				path = entry.OriginalPath
			}
			if path == "/"+selector.Path && entry.ID > 0 {
				return map[string]any{"changeTrackingId": entry.ID, "iterationContext": map[string]int{"firstComparingIteration": 0, "secondComparingIteration": iteration}}, nil
			}
		}
		if changes.NextTop == 0 {
			return nil, fmt.Errorf("%w: path is not in captured PR diff", comments.ErrInvalid)
		}
		if changes.NextSkip <= skip {
			return nil, fmt.Errorf("native Azure iteration cursor did not advance")
		}
		skip = changes.NextSkip
	}
}
