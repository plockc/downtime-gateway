package namespace

import (
	"fmt"
	"strings"

	"github.com/plockc/gateway/exec"
	"github.com/plockc/gateway/resource"
)

type NS string

var _ resource.Resource = NS("")

func (ns NS) Id() string {
	return string(ns)
}

func (ns NS) String() string {
	return ns.Id()
}

func (ns NS) Delete() error {
	_, _, err := exec.ExecLine("ip netns del " + string(ns))
	return err
}

func (ns NS) List() ([]string, error) {
	_, out, err := exec.ExecLine("ip netns list")
	if err != nil {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

func (ns NS) Create() error {
	_, _, err := exec.ExecLine("ip netns add " + string(ns))
	return err
}

func (ns NS) Clear() error {
	return fmt.Errorf("Clear not implemented")
}
