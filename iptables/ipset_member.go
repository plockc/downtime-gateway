package iptables

import (
	"fmt"
	"strings"

	"github.com/plockc/gateway/address"
	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/resource"
)

type Member struct {
	address.MAC
	// avoids embedding so this can be assured to implement the Resource interface
	IPSet IPSet
	resource.FailUnimplementedMethods
}

var _ resource.Resource = &Member{}

func NewMember(ipSet IPSet, mac address.MAC) Member {
	return Member{MAC: mac, IPSet: ipSet}
}

func (m Member) Id() string {
	return m.MAC.String()
}

func (m Member) String() string {
	return m.Id()
}

func (m Member) Create() error {
	return m.IPSet.Runner.Line("ipset add " + m.IPSet.Id() + " " + m.MAC.String())
}

func (m Member) Delete() error {
	return m.IPSet.Runner.Line("ipset del " + m.IPSet.Id() + " " + m.MAC.String())
}

func (m Member) List() ([]string, error) {
	run := m.IPSet.Runner
	err := run.Line("ipset save -sorted " + m.IPSet.Id())
	if err != nil {
		return nil, fmt.Errorf("failed to list members of ipset '%s': %w", m.IPSet.Id(), err)
	}
	elems := funcs.Keep(strings.Split(run.LastOut(), "\n"), func(s string) bool {
		return strings.HasPrefix(s, "add ")
	})
	macIds := funcs.Map(elems, func(s string) string {
		return strings.TrimPrefix(s, "add "+m.IPSet.Id()+" ")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse macs: %w", err)
	}
	return macIds, nil
}

func (m Member) Clear() error {
	return m.IPSet.Runner.Line("ipset flush " + m.IPSet.Id())
}
