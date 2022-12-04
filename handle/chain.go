package handle

import (
	"reflect"

	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func NewChain(ids ...string) (iptables.Chain, error) {
	table, err := NewTable(ids...)
	if err != nil {
		return iptables.Chain{}, err
	}
	switch len(ids) {
	case 3:
		return iptables.NewChain(table, ""), nil
	default:
		return iptables.NewChain(table, ids[3]), nil
	}
}

var Chains = Resources{
	Name: "IPTable Chain",
	Factory: func(bodyIgnored []byte, ids ...string) (resource.Resource, error) {
		chain, err := NewChain(ids...)
		if err != nil {
			return nil, err
		}
		return chain.ChainResource(), nil
	},
	Relationships: map[string]Resources{
		"rules": Rules,
	},
	T:       reflect.TypeOf(""),
	Allowed: []Allowed{LIST_ALLOWED, DELETE_ALLOWED, GET_ALLOWED, UPSERT_ALLOWED},
}
