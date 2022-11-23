package exec

import (
	"fmt"
	"os/exec"
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
			c.String(), t.ExitCode(), string(outString),
		)
	case error:
		return 1, outString, fmt.Errorf(
			"failed to run `%s`: %w", c.String(), err,
		)
	}
	return 0, outString, nil
}
