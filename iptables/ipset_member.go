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
	IPSet
}

func (member Member) String() string {
	return member.IPSet.String() + ":member[" + member.Name + "]"
}

func (m Member) MemberResource() MemberRes {
	return MemberRes{Member: m}
}

func NewMember(ipSet IPSet, mac address.MAC) Member {
	return Member{MAC: mac, IPSet: ipSet}
}

var _ resource.Resource = &MemberRes{}

type MemberRes struct {
	Member
	resource.FailUnimplementedMethods
}

func NewMemberResource(m Member) MemberRes {
	return MemberRes{Member: m}
}

func (m MemberRes) Id() string {
	return m.MAC.String()
}

func (m MemberRes) Create() error {
	return m.Runner().RunLine("ipset add " + m.IPSet.Name + " " + m.MAC.String())
}

func (m MemberRes) Delete() error {
	return m.Runner().RunLine("ipset del " + m.IPSet.Name + " " + m.MAC.String())
}

func (m MemberRes) List() ([]string, error) {
	run := m.Runner()
	setName := m.IPSet.Name
	err := run.RunLine("ipset save -sorted " + setName)
	if err != nil {
		return nil, fmt.Errorf("failed to list members of ipset '%s': %w", setName, err)
	}
	elems := funcs.Keep(strings.Split(run.LastOut(), "\n"), func(s string) bool {
		return strings.HasPrefix(s, "add ")
	})
	macIds := funcs.Map(elems, func(s string) string {
		return strings.TrimPrefix(s, "add "+setName+" ")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse macs: %w", err)
	}
	return macIds, nil
}

func (m MemberRes) Clear() error {
	return m.Runner().RunLine("ipset flush " + m.IPSet.Name)
}
