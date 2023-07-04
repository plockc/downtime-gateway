package handle

import (
	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func ChainChainedFactory(chain *iptables.Chain) ChainedFactory {
	return func() (ChainedFactory, Factory) {
		factory := func(chainName string) (resource.Resource, error) {
			(*chain).Name = chainName
			return chain.ChainResource(), nil
		}
		return TableChainedFactory(&chain.Table), factory
	}
}

var Chains = Resources{
	Name: "IPTable Chain",
	ChainedFactory: func() (ChainedFactory, Factory) {
		chain := iptables.Chain{}
		return ChainChainedFactory(&chain)()
	},
	Relationships: map[string]Resources{
		"rules": Rules,
	},
	Allowed: []Allowed{LIST_ALLOWED, DELETE_ALLOWED, GET_ALLOWED, UPSERT_ALLOWED},
}
