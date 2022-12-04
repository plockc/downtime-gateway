package handle

import (
	"reflect"
	"strconv"

	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

var Rules = Resources{
	Name: "IPTables Rules",
	Factory: func(body []byte, ids ...string) (resource.Resource, error) {
		chain, err := NewChain(ids...)
		if err != nil {
			return nil, err
		}
		// prior ids are version, namespace, table, chain
		switch len(ids) {
		case 4:
			return iptables.NewRule(chain).RuleResource(), nil
		default:
			rule := iptables.NewRule(chain)
			id, err := strconv.ParseUint(ids[4], 16, 32)
			if err != nil {
				return iptables.Rule{}.RuleResource(), err
			}
			rule.Id = uint32(id)
			return rule.RuleResource(), nil
		}
	},
	T:       reflect.TypeOf(""),
	Allowed: []Allowed{LIST_ALLOWED, UPSERT_ALLOWED, GET_ALLOWED, DELETE_ALLOWED},
}
