package handle

import (
	"reflect"

	"github.com/plockc/gateway/resource"
)

var Namespaces = Resources{
	Name: "Network Namespace",
	Factory: func(bodyIgnored []byte, ids ...string) (resource.Resource, error) {
		switch len(ids) {
		case 0, 1:
			return resource.NewNS("").Resource(), nil
		default:
			return resource.NewNS(ids[0]).Resource(), nil
		}
	},
	T: reflect.TypeOf([]string{}),
	Relationships: map[string]Resources{
		"iptables": Tables,
		"ipsets":   IPSets,
	},
	Allowed: []Allowed{GET_ALLOWED, LIST_ALLOWED},
}
