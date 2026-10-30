package runconfig

import "github.com/flanksource/commons/merge"

var optionPolicy = merge.Policy{Replace: []any{[]string(nil), (*bool)(nil), (*int)(nil), (*string)(nil)}}

func (o Test) Merge(override Test) Test { return merge.Apply(o, override, optionPolicy) }

func (o Lint) Merge(override Lint) Lint { return merge.Apply(o, override, optionPolicy) }
