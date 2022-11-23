package iptables

import (
	"strings"

	"github.com/plockc/gateway/namespace"
	"github.com/plockc/gateway/resource"
	"github.com/plockc/gateway/runner"
)

type IPSet struct {
	resource.Named `json:"-"`
	Runner         *runner.Runner `json:"-"`
	resource.FailUnimplementedMethods
}

var _ resource.Resource = IPSet{}

func NewIPSet(ns namespace.NS, name string) IPSet {
	return IPSet{Named: resource.Named(name), Runner: runner.NamespacedRunner(ns)}
}

func (ipSet IPSet) Delete() error {
	return ipSet.Runner.Line("ipset destroy " + ipSet.Id())
}

func (ipSet IPSet) Create() error {
	return ipSet.Runner.Line(
		"ipset -N "+ipSet.Id()+" hash:mac",
		"ipset -N -exist "+ipSet.Id()+"-builder hash:mac",
		"ipset swap "+ipSet.Id()+"-builder "+ipSet.Id(),
		"ipset destroy "+ipSet.Id()+"-builder",
	)
}

func (ipSet IPSet) List() ([]string, error) {
	if err := ipSet.Runner.Line("ipset list -n"); err != nil {
		return nil, err
	}
	return strings.Split(ipSet.Runner.LastOut(), "\n"), nil
}

func (ipSet IPSet) Match() string {
	return "-m set --match-set " + ipSet.String() + " src"
}
