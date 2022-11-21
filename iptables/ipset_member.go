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
}

var _ resource.Resource[address.MAC] = &Member{}

func NewMember(ipSet IPSet, mac address.MAC) Member {
	return Member{MAC: mac, IPSet: ipSet}
}

func (m Member) Id() address.MAC {
	return m.MAC
}

func (m Member) String() string {
	return m.MAC.String()
}

func (m Member) Create() error {
	return m.IPSet.Runner.Line("ipset add " + m.IPSet.Name + " " + m.MAC.String())
}

func (m Member) Delete() error {
	return m.IPSet.Runner.Line("ipset del " + m.IPSet.Name + " " + m.MAC.String())
}

func (m Member) List() ([]address.MAC, error) {
	run := m.IPSet.Runner
	err := run.Line("ipset save -sorted " + m.IPSet.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list members of ipset %s: %w", m.IPSet.Name, err)
	}
	elems := funcs.Keep(strings.Split(run.LastOut(), "\n"), func(s string) bool {
		return strings.HasPrefix(s, "add ")
	})
	elems = funcs.Map(elems, func(s string) string {
		return strings.TrimPrefix(s, "add "+m.IPSet.Name+" ")
	})
	macs, err := funcs.MapErrable(elems, func(s string) (address.MAC, error) {
		return address.MACFromString(s)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse macs: %w", err)
	}
	return macs, nil
}
