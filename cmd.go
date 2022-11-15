package gateway

import (
	"fmt"
	"os/exec"
	"strings"
)

type Cmds [][]string

func (cmds *Cmds) AddCmdLine(cmd ...string) []int {
	return cmds.Add(Multiline(cmd).Split()...)
}

func (cmds *Cmds) Add(cmd ...[]string) []int {
	offset := len(*cmds)
	idxs := []int{}
	for _, c := range cmd {
		*cmds = append(*cmds, c)
		idxs = append(idxs, offset)
		offset++
	}
	return idxs
}

func RunCmdLine(cmd string) (string, error) {
	outs, err := Run(strings.Split(cmd, " "))
	return outs[0], err
}

func RunCmdLines(cmds ...string) ([]string, error) {
	return Run(Multiline(cmds).Split()...)
}

func Run(cmds ...[]string) ([]string, error) {
	outs := []string{}
	for _, cmd := range cmds {
		c := exec.Command(cmd[0], cmd[1:]...)
		out, err := c.CombinedOutput()
		outs = append(outs, string(out))
		switch t := err.(type) {
		case *exec.ExitError:
			return outs, fmt.Errorf(
				"`%s` failed with exit code %d: %s",
				c.String(), t.ExitCode(), string(t.Stderr),
			)
		case error:
			return outs, fmt.Errorf(
				"failed to run `%s`: %w", c.String(), err,
			)
		}
	}

	return outs, nil
}

func (cmds Cmds) CmdLines() []string {
	return FromJoin(cmds)
}

// can be used with Multiline to Print out
func (cmds Cmds) Debug(outs []string) []string {
	cmdlines := []string{}
	for i, out := range outs {
		cmdline := cmds[i]
		cmdlines = append(cmdlines, strings.Join(cmdline, " "), out)
	}
	return cmdlines
}
