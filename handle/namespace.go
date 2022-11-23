package handle

import (
	"reflect"

	"github.com/plockc/gateway/namespace"
	"github.com/plockc/gateway/resource"
)

var Namespaces = Resources{
	Name: "Network Namespace",
	Factory: func(bodyIgnored []byte, ids ...string) (resource.Resource, error) {
		ns := namespace.NS("")
		switch len(ids) {
		case 0, 1:
			return ns, nil
		default:
			return namespace.NS(ids[0]), nil
		}
	},
	T: reflect.TypeOf([]string{}),
	Relationships: map[string]Resources{
		"ipsets": IPSets,
	},
	Allowed: []Allowed{GET_ALLOWED, LIST_ALLOWED},
}
