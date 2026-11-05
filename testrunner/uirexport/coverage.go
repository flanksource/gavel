package uirexport

import (
	"cmp"
	"fmt"
	"slices"

	"golang.org/x/tools/cover"
)

// readProfiles merges go test -coverprofile files into module-relative blocks. A block listed by
// several profiles, as concurrent batches or -coverpkg produce, has its counts summed. Blocks of files
// outside the module, such as a nested module's, are dropped: uir serves only the module it indexed.
func readProfiles(paths []string, m *module) ([]Block, error) {
	merged := map[Block]int64{}
	for _, path := range paths {
		profiles, err := cover.ParseProfiles(path)
		if err != nil {
			return nil, fmt.Errorf("read coverage profile %s: %w", path, err)
		}
		for _, profile := range profiles {
			file := m.profileFile(profile.FileName)
			if file == "" {
				continue
			}
			for _, block := range profile.Blocks {
				key := Block{Path: file, StartLine: block.StartLine, StartCol: block.StartCol, EndLine: block.EndLine, EndCol: block.EndCol, Statements: block.NumStmt}
				merged[key] += int64(block.Count)
			}
		}
	}
	blocks := make([]Block, 0, len(merged))
	for block, count := range merged {
		block.Count = count
		blocks = append(blocks, block)
	}
	slices.SortFunc(blocks, func(left, right Block) int {
		return cmp.Or(cmp.Compare(left.Path, right.Path), cmp.Compare(left.StartLine, right.StartLine), cmp.Compare(left.StartCol, right.StartCol))
	})
	return blocks, nil
}
