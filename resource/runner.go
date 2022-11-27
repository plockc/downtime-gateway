package resource

import (
	"strconv"
	"strings"

	"github.com/plockc/gateway/exec"
	"github.com/plockc/gateway/multiline"
)

type Runner struct {
	// prevent Id and String from promoting so
	// can embed Runner without conflicting with
	// an embedded resource.Named
	NS
	Results []Result
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

func NamespacedRunner(ns NS) *Runner {
	return &Runner{NS: ns}
}

func (r *Runner) WrapCmd(cmd []string) []string {
	if r == nil || r.NSName() == "" {
		return cmd
	}
	return append(
		[]string{"ip", "netns", "exec", string(r.NSName())}, cmd...,
	)
}

func (r *Runner) WrapCmdLine(cmd string) string {
	if r == nil || r.NSName() == "" {
		return cmd
	}
	return "ip netns exec " + string(r.NSName()) + " " + cmd
}

func (r *Runner) WrapCmdLines(cmds []string) []string {
	return multiline.Multiline(cmds).Map(r.WrapCmdLine)
}

func (r *Runner) Last() Result {
	if r == nil {
		return Result{
			Code: 1,
			Out:  "internal error, result tracking not initialized",
		}
	}
	return r.Results[len(r.Results)-1]
}

func (r *Runner) LastOut() string {
	return r.Last().Out
}

// LastFund is basically defers getting the result, useful when the last command is inside a Do(...)
// needs a pointer as otherwise the array state (likely empty) when calling LastOutFunc is copied
// instead of using the updated Results array pointer maintained by runner
func (r *Runner) LastOutFunc() func() string {
	return func() string {
		return r.Last().Out
	}
}

func (r *Runner) String() string {
	out := []string{}
	for _, d := range r.Results {
		out = append(out, d.String())
	}
	return strings.Join(out, "\n")
}

func (r *Runner) Line(cmds ...string) error {
	return r.Run(multiline.Multiline(cmds).Split()...)
}

func (r *Runner) Run(cmds ...[]string) error {
	for _, cmd := range cmds {
		// namspace the command if we're in a namespace
		if r.NSName() != "" {
			cmd = r.WrapCmd(cmd)
		}
		code, out, err := exec.Exec(cmd)
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
		return r.Run(multiline.Multiline(cmds).Split()...)
	}
}

// will appends the outputs of cmds to the Runer
// appending so can debug the full chain of commands when used multiple times in Do(...)
func (r *Runner) Func(cmds ...[]string) func() error {
	return func() error {
		return r.Run(cmds...)
	}
}
