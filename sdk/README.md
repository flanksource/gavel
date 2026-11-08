# gavel sdk

`github.com/flanksource/gavel/sdk` holds the small, dependency-free contracts other tools use to read what running gavel processes are doing. It is a separate Go module, so importing it does not pull in gavel's dependencies.

## status

Each running `gavel test`, `gavel lint` and `gavel fixtures` publishes a `status.Status` to `<cache>/status/<pid>.json`. The status carries test counts, the task-group progress the test UI shows, the tasks running now, and the slots the process holds or waits for. After 30 seconds the process also serves it on `<cache>/status/<pid>.sock` (`GET /status`, and `GET /status/stream` as server-sent events). `<cache>` is `status.CacheDir()`: `$XDG_CACHE_HOME/gavel`, or the OS user cache directory's `gavel`.

List every live gavel process:

```go
package main

import (
	"fmt"

	"github.com/flanksource/gavel/sdk/status"
)

func main() {
	cache, err := status.CacheDir()
	if err != nil {
		panic(err)
	}
	statuses, err := status.List(cache)
	if err != nil {
		panic(err)
	}
	for _, s := range statuses {
		fmt.Printf("%d %s groups=%d slots=%d\n", s.PID, s.Command, len(s.Groups), len(s.Slots))
	}
}
```

Follow one process until it exits. `Watch` polls the file, then switches to the socket's stream once the process serves one:

```go
watcher, err := status.Watch(ctx, status.WatchOptions{Cache: cache, PID: pid})
if err != nil {
	return err
}
for s := range watcher.Updates() {
	for _, group := range s.Groups {
		fmt.Printf("%s %d/%d failed=%d\n", group.Name, group.Summary.Done(), group.Summary.Total, group.Summary.Failed)
	}
}
return watcher.Err()
```

Release: tag `sdk/vX.Y.Z` on the gavel repository; modules depending on gavel resolve the sdk by that tag.
