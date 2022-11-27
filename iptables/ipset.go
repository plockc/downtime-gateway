package iptables

import (
	"strings"

	"github.com/plockc/gateway/resource"
)

type IPSet struct {
	Name string
	resource.NS
}

var _ resource.Resource = IPSetRes{}

func NewIPSet(ns resource.NS, name string) IPSet {
	return IPSet{Name: name, NS: ns}
}

func (ipSet IPSet) Resource() resource.Resource {
	return IPSetRes{IPSet: ipSet}
}

func (ipSet IPSet) Match() string {
	return "-m set --match-set " + ipSet.Name + " src"
}

func (ipSet IPSet) String() string {
	return ipSet.NS.String() + ":ipSet[" + ipSet.Name + "]"
}

type IPSetRes struct {
	resource.FailUnimplementedMethods
	IPSet
}

func (ipSet IPSetRes) Id() string {
	return ipSet.Name
}

func (ipSet IPSetRes) Delete() error {
	return ipSet.Runner().Line("ipset destroy " + ipSet.Id())
}

func (ipSet IPSetRes) Create() error {
	return ipSet.Runner().Line(
		"ipset -N "+ipSet.Id()+" hash:mac",
		"ipset -N -exist "+ipSet.Id()+"-builder hash:mac",
		"ipset swap "+ipSet.Id()+"-builder "+ipSet.Id(),
		"ipset destroy "+ipSet.Id()+"-builder",
	)
}

func (ipSet IPSetRes) List() ([]string, error) {
	runner := ipSet.Runner()
	if err := runner.Line("ipset list -n"); err != nil {
		return nil, err
	}
	return strings.Split(runner.LastOut(), "\n"), nil
}
