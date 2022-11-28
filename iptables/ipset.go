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

func (ipSet IPSet) Match() string {
	return "-m set --match-set " + ipSet.Name + " src"
}

func (ipSet IPSet) String() string {
	return ipSet.NS.String() + ":ipSet[" + ipSet.Name + "]"
}

func (ipSet IPSet) IPSetResource() IPSetRes {
	return IPSetRes{IPSet: ipSet}
}

type IPSetRes struct {
	resource.FailUnimplementedMethods
	IPSet
}

func (ipSet IPSetRes) Id() string {
	return ipSet.Name
}

func (ipSet IPSetRes) Delete() error {
	return ipSet.Runner().RunLine("ipset destroy " + ipSet.Id())
}

func (ipSet IPSetRes) Create() error {
	return ipSet.Runner().RunLine("ipset -N " + ipSet.Id() + " hash:mac")
}

func (ipSet IPSetRes) List() ([]string, error) {
	runner := ipSet.Runner()
	if err := runner.RunLine("ipset list -n"); err != nil {
		return nil, err
	}
	return strings.Split(runner.LastOut(), "\n"), nil
}
