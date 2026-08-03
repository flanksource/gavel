package main

import "github.com/flanksource/clicky/entity"

// clicky only wires an Options struct's long help into cobra when it satisfies
// entity.Help exactly (`Help() api.Textable`). A near-miss signature such as
// `Help() string` or `Help() api.Text` compiles fine and is silently ignored,
// so `--help` loses its description and examples with no error anywhere. These
// assertions turn that into a compile failure.
var (
	_ entity.Help = CommitOptions{}
	_ entity.Help = PRListOptions{}
	_ entity.Help = PRStatusOptions{}
	_ entity.Help = ProcListOptions{}
	_ entity.Help = ProcLogsOptions{}
	_ entity.Help = ProcRestartOptions{}
	_ entity.Help = ProcRunOptions{}
	_ entity.Help = ProcStartOptions{}
	_ entity.Help = ProcStatusOptions{}
	_ entity.Help = ProcStopOptions{}
	_ entity.Help = RepomapGetOptions{}
	_ entity.Help = RepomapViewOptions{}
	_ entity.Help = SSHInstallOptions{}
	_ entity.Help = ServeOptions{}
	_ entity.Help = StatusOptions{}
	_ entity.Help = SystemInstallOptions{}
	_ entity.Help = SystemStartOptions{}
	_ entity.Help = SystemStatusOptions{}
	_ entity.Help = SystemStopOptions{}
	_ entity.Help = SystemUninstallOptions{}
	_ entity.Help = UIServeOptions{}
	_ entity.Help = fixturesOutlineOptions{}
	_ entity.Help = testANSIOptions{}
	_ entity.Help = testHistoryOptions{}
	_ entity.Help = testOutlineOptions{}
)
