package resource

import (
	"fmt"
	"strings"

	"github.com/plockc/gateway/exec"
)

type NS struct {
	Name string
}

func NewNS(name string) NS {
	return NS{Name: name}
}

func (ns NS) Resource() Resource {
	return NSRes{Name: ns.Name}
}

func (ns NS) Runner() *Runner {
	return &Runner{NS: ns}
}

func (ns NS) NSName() string {
	return ns.Name
}

func (ns NS) String() string {
	return "namespace[" + ns.Name + "]"
}

type NSRes struct {
	Name string
	FailUnimplementedMethods
}

var _ Resource = NSRes{}

func (ns NSRes) Id() string {
	return ns.Name
}

func (ns NSRes) Delete() error {
	_, _, err := exec.ExecLine("ip netns del " + ns.Id())
	return err
}

func (ns NSRes) List() ([]string, error) {
	_, out, err := exec.ExecLine("ip netns list")
	if err != nil {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

func (ns NSRes) Create() error {
	_, _, err := exec.ExecLine("ip netns add " + ns.Id())
	return err
}

func (ns NSRes) Clear() error {
	return fmt.Errorf("Clear not implemented")
}
