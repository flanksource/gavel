package procfile

import "fmt"

func resolveProcessSelection(entries []Entry, opts Options) ([]Entry, map[string]bool, error) {
	if len(opts.Names) > 0 && len(opts.StartNames) > 0 {
		return nil, nil, fmt.Errorf("supervisor cannot combine Names and StartNames")
	}

	registered, err := Select(entries, opts.Names)
	if err != nil {
		return nil, nil, err
	}

	names := opts.StartNames
	if len(names) == 0 {
		names = opts.Names
	}
	if len(names) == 0 {
		return registered, nil, nil
	}

	selected, err := Select(entries, names)
	if err != nil {
		return nil, nil, err
	}
	startNames := make(map[string]bool, len(selected))
	for _, entry := range selected {
		startNames[entry.Name] = true
	}
	return registered, startNames, nil
}
