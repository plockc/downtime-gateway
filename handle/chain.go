package handle

import (
	"reflect"

	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

var Chains = Resources{
	Name: "IPTable Chain",
	Factory: func(bodyIgnored []byte, ids ...string) (resource.Resource, error) {
		table, err := NewTable(ids...)
		if err != nil {
			return nil, err
		}
		switch len(ids) {
		default:
			return iptables.NewChain(table, ids[2]).Resource(), nil
		}
	},
	T:       reflect.TypeOf(""),
	Allowed: []Allowed{LIST_ALLOWED},
}
