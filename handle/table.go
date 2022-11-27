package handle

import (
	"fmt"
	"reflect"

	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func NewTable(ids ...string) (iptables.Table, error) {
	switch len(ids) {
	case 0, 1:
		return iptables.Table{}, fmt.Errorf("missing version and/or namespace")
	case 2:
		return iptables.NewTable(resource.NewNS(ids[1]), "filter"), nil
	default:
		return iptables.NewTable(resource.NewNS(ids[1]), ids[2]), nil
	}
}

var Tables = Resources{
	Name: "IP Table",
	Factory: func(bodyIgnored []byte, ids ...string) (resource.Resource, error) {
		table, err := NewTable(ids...)
		if err != nil {
			return nil, err
		}
		return table.Resource(), nil
	},
	T: reflect.TypeOf(""),
	Relationships: map[string]Resources{
		"chains": Chains,
	},
	Allowed: []Allowed{LIST_ALLOWED},
}
