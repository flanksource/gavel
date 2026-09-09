package types

import "testing"

func TestParseRelationKind(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    RelationKind
		wantErr bool
	}{
		{name: "empty defaults to related_to", input: "", want: RelationRelatedTo},
		{name: "snake case depends_on", input: "depends_on", want: RelationDependsOn},
		{name: "kebab case depends-on", input: "depends-on", want: RelationDependsOn},
		{name: "mixed case is normalized", input: "Depends-On", want: RelationDependsOn},
		{name: "snake case related_to", input: "related_to", want: RelationRelatedTo},
		{name: "kebab case related-to", input: "related-to", want: RelationRelatedTo},
		{name: "blocks is derived and cannot be written", input: "blocks", wantErr: true},
		{name: "unknown relation is rejected", input: "duplicates", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRelationKind(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseRelationKind(%q) = %q, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRelationKind(%q): %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("ParseRelationKind(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// blocks is the reverse projection of depends_on, so it must never appear in
// the set a caller is allowed to write.
func TestLinkableRelationsExcludeBlocks(t *testing.T) {
	for _, relation := range LinkableRelations() {
		if relation == RelationBlocks {
			t.Fatal("LinkableRelations() includes the derived blocks relation")
		}
	}
	if len(LinkableRelations()) != 2 {
		t.Fatalf("LinkableRelations() = %v, want depends_on and related_to", LinkableRelations())
	}
}
