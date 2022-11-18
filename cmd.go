package gateway

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func ExecLine(cmd string) (int, string, error) {
	return Exec(strings.Split(cmd, " "))
}

// Exec will trim the trailing new line
// TODO: add variadic option parameter to keep trailing newline
func Exec(cmd []string) (int, string, error) {
	c := exec.Command(cmd[0], cmd[1:]...)
	out, err := c.CombinedOutput()
	outString := ""
	if out != nil {
		outString = string(out)
	}
	outString = strings.TrimRightFunc(outString, func(r rune) bool {
		return r == '\n'
	})
	switch t := err.(type) {
	case *exec.ExitError:
		return t.ExitCode(), outString, fmt.Errorf(
			"`%s` failed with exit code %d: %s",
			c.String(), t.ExitCode(), string(t.Stderr),
		)
	case error:
		return 1, outString, fmt.Errorf(
			"failed to run `%s`: %w", c.String(), err,
		)
	}
	return 0, outString, nil
}

type Result struct {
	Cmd  []string
	Out  string
	Code int
}

func (d Result) String() string {
	if d.Code != 0 {
		return "[" + strconv.Itoa(d.Code) + "] " + strings.Join(d.Cmd, " ") + "\n" + d.Out
	}
	return strings.Join(d.Cmd, " ") + "\n" + d.Out
}

type Runner struct {
	NS
	Results []Result
}

func NamespacedRunner(ns NS) *Runner {
	return &Runner{NS: ns}
}

func (r Runner) Last() Result {
	return r.Results[len(r.Results)-1]
}

func (r Runner) LastOut() string {
	return r.Results[len(r.Results)-1].Out
}

// LastFund is basically defers getting the result, useful when the last command is inside a Do(...)
// needs a pointer as otherwise the array state (likely empty) when calling LastOutFunc is copied
// instead of using the updated Results array pointer maintained by runner
func (r *Runner) LastOutFunc() func() string {
	return func() string {
		return r.Results[len(r.Results)-1].Out
	}
}

func (r Runner) String() string {
	out := []string{}
	for _, d := range r.Results {
		out = append(out, d.String())
	}
	return strings.Join(out, "\n")
}

func (r *Runner) Line(cmds ...string) error {
	return r.Run(Multiline(cmds).Split()...)
}

func (r *Runner) Run(cmds ...[]string) error {
	for _, cmd := range cmds {
		// namspace the command if we're in a namespace
		if r.NS != "" {
			cmd = r.NS.WrapCmd(cmd)
		}
		code, out, err := Exec(cmd)
		r.Results = append(r.Results, Result{Cmd: cmd, Out: out, Code: code})
		if err != nil {
			return err
		}
	}

	return nil
}

// will appends the outputs of cmds to the Runer
// appending so can debug the full chain of commands when used multiple times in Do(...)
func (r *Runner) LineFunc(cmds ...string) func() error {
	return func() error {
		return r.Run(Multiline(cmds).Split()...)
	}
}

// will appends the outputs of cmds to the Runer
// appending so can debug the full chain of commands when used multiple times in Do(...)
func (r *Runner) Func(cmds ...[]string) func() error {
	return func() error {
		return r.Run(cmds...)
	}
}
