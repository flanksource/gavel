package verify

type Adapter interface {
	Name() string
	BuildVerifyArgs(prompt, model, schemaFile string, debug bool) []string
	BuildFixArgs(model, prompt string, patchOnly bool) []string
	ParseResponse(raw string) (VerifyResult, error)
	PostExecute(raw string)
	ListModels() ([]string, error)
}
