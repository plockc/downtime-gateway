package handle

import (
	"github.com/plockc/gateway/resource"
)

var Namespaces = Resources{
	Name: "Network Namespace",
	Factory: func(ids ...string) (resource.Resource, error) {
		switch len(ids) {
		case 0, 1:
			return resource.NewNS("").NSResource(), nil
		default:
			return resource.NewNS(ids[0]).NSResource(), nil
		}
	},
	Relationships: map[string]Resources{
		"iptables": Tables,
		"ipsets":   IPSets,
	},
	Allowed: []Allowed{GET_ALLOWED, LIST_ALLOWED},
}
