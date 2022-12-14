package handle

import (
	"fmt"

	"github.com/plockc/gateway/address"
	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func NewIPSet(ids ...string) (iptables.IPSet, error) {
	switch len(ids) {
	case 0, 1:
		return iptables.IPSet{}, fmt.Errorf("missing version and/or namespace")
	case 2:
		return iptables.NewIPSet(resource.NewNS(ids[1]), ""), nil
	default:
		return iptables.NewIPSet(resource.NewNS(ids[1]), ids[2]), nil
	}
}

var IPSets = Resources{
	Name: "IP Set",
	Factory: func(ids ...string) (resource.Resource, error) {
		ipSet, err := NewIPSet(ids...)
		if err != nil {
			return nil, err
		}
		return ipSet.IPSetResource(), nil
	},
	Relationships: map[string]Resources{
		"members": IPSetMembers,
	},
	Allowed: []Allowed{GET_ALLOWED, LIST_ALLOWED, DELETE_ALLOWED, UPSERT_ALLOWED},
}

var IPSetMembers = Resources{
	Name: "IP Set Member",
	Factory: func(ids ...string) (resource.Resource, error) {
		ipSet, err := NewIPSet(ids...)
		if err != nil {
			return nil, err
		}
		switch len(ids) {
		case 2:
			return nil, fmt.Errorf("missing ipset name")
		case 3:
			return iptables.NewMember(ipSet, address.MAC{}).MemberResource(), nil
		default:
			mac, err := address.MACFromString(ids[3])
			return iptables.NewMember(ipSet, mac).MemberResource(), err
		}
	},
	Allowed: []Allowed{GET_ALLOWED, LIST_ALLOWED, DELETE_ALLOWED, UPSERT_ALLOWED},
}
