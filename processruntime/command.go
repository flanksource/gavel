package processruntime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"sync"

	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/slots"
)

type Command struct {
	ctx         context.Context
	slots       slots.Limits
	Dir         string
	Env         []string
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader
	Process     *cexec.Process
	ConfigError error
}

func NewCommand(ctx context.Context, options Options, binary string, args ...string) *Command {
	p := cexec.NewExec(binary, args...).WithContext(ctx).WithProcessGroup()
	return &Command{Process: p, ConfigError: options.Apply(p), ctx: ctx, slots: options.Slots}
}

func (c *Command) Run() (err error) {
	if c.ConfigError != nil {
		return c.ConfigError
	}
	c.Process.WithCwd(c.Dir).Stream(c.Stdout, c.Stderr)
	for _, entry := range c.Env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			return errors.New("invalid command environment entry")
		}
		c.Process.WithEnv(map[string]string{key: value})
	}
	if c.Stdin != nil {
		c.Process.WithStdin(c.Stdin)
	}
	lease, err := slots.Acquire(c.ctx, slots.Request{Limits: c.slots, WorkDir: c.Dir, Command: strings.Join(append([]string{c.Process.Cmd}, c.Process.Args...), " ")})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := lease.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	result := c.Process.WithEnv(lease.Env()).Run().Result()
	var native *exec.ExitError
	if errors.As(result.Error, &native) {
		return native
	}
	return result.Error
}

func (c *Command) CombinedOutput() ([]byte, error) {
	output := &commandOutput{}
	c.Stdout, c.Stderr = output, output
	err := c.Run()
	return output.Bytes(), err
}

type commandOutput struct {
	sync.Mutex
	bytes.Buffer
}

func (o *commandOutput) Write(data []byte) (int, error) {
	o.Lock()
	defer o.Unlock()
	return o.Buffer.Write(data)
}
