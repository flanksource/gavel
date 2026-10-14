package verify

func runtimeSchema() map[string]any {
	nice := intProp("Child niceness adjustment, from -20 to 19. Increasing priority requires OS permission.")
	nice["minimum"], nice["maximum"] = -20, 19
	cpu := numberProp("Aggregate sustained CPU percentage across descendants. 100 is one core; 0 disables the cap. Sampled every 2s, terminating after 18 consecutive high samples.")
	cpu["minimum"] = 0
	return object("Process settings inherited by all linter and test types. Within each config file, global defaults precede category and type overrides; later files and explicit caller options win.", map[string]any{
		"env":     mapObject("Child environment variables. Keys merge; an empty value clears an inherited value.", stringProp("Environment value.")),
		"nice":    nice,
		"cpu":     cpu,
		"mem":     stringProp("Aggregate RSS cap, e.g. 2GiB or 512MB. Empty disables the cap. Includes descendants and is sampled immediately, then every 2s."),
		"timeout": stringProp("Positive Go duration for each invocation, e.g. 5m. Empty disables the inherited per-invocation deadline. Separate from the whole-run deadline."),
	})
}
