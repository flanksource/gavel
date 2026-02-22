package verify

import "strings"

type Gemini struct{}

func (Gemini) Name() string { return "gemini" }

func (Gemini) BuildVerifyArgs(prompt, model, _ string, debug bool) []string {
	args := []string{"-p", prompt, "--output-format", "json"}
	if model != "" && model != "gemini" {
		args = append(args, "-m", model)
	}
	if debug {
		args = append(args, "--debug")
	}
	return args
}

func (Gemini) BuildFixArgs(model, prompt string, _ bool) []string {
	args := []string{"-p", prompt}
	if model != "" && model != "gemini" {
		args = append(args, "-m", model)
	}
	return args
}

func (Gemini) ParseResponse(raw string) (VerifyResult, error) {
	if result, ok := tryUnmarshalResult(raw); ok {
		return result, nil
	}
	cleaned := strings.TrimSpace(stripMarkdownFences(raw))
	if result, ok := tryUnmarshalResult(cleaned); ok {
		return result, nil
	}
	return VerifyResult{}, parseError(raw)
}

func (Gemini) PostExecute(string) {}

func (Gemini) ListModels() ([]string, error) {
	key := getEnv("GEMINI_API_KEY", "GOOGLE_API_KEY")
	if key == "" {
		return nil, nil
	}
	return fetchModelIDs("https://generativelanguage.googleapis.com/v1beta/models?key="+key, "", "", "")
}
