package iptables

import (
	"strings"

	"github.com/plockc/gateway/namespace"
	"github.com/plockc/gateway/resource"
	"github.com/plockc/gateway/runner"
)

type IPSet struct {
	Name   string
	Runner *runner.Runner
}

var _ resource.Resource[string] = IPSet{}

func NewIPSet(name string, ns namespace.NS) IPSet {
	return IPSet{Name: name, Runner: runner.NamespacedRunner(ns)}
}

func (ipSet IPSet) Id() string {
	return ipSet.Name
}

func (ipSet IPSet) String() string {
	return ipSet.Id()
}

func (ipSet IPSet) LastResult() runner.Result {
	return ipSet.Runner.Last()
}

func (ipSet IPSet) Delete() error {
	return ipSet.Runner.Line("ipset destroy " + ipSet.Name)
}

func (ipSet IPSet) Create() error {
	return ipSet.Runner.Line(
		"ipset -N "+ipSet.Name+" hash:mac",
		"ipset -N -exist "+ipSet.Name+"-builder hash:mac",
		"ipset swap "+ipSet.Name+"-builder "+ipSet.Name,
		"ipset destroy "+ipSet.Name+"-builder",
	)
}

func (ipSet IPSet) List() ([]string, error) {
	if err := ipSet.Runner.Line("ipset list -n"); err != nil {
		return nil, err
	}
	return strings.Split(ipSet.Runner.LastOut(), "\n"), nil
}

func (ipSet IPSet) Clear() error {
	return ipSet.Runner.Line("ipset flush " + ipSet.Name)
}

func (ipSet IPSet) Match() string {
	return "-m set --match-set " + ipSet.String() + " src"
}
