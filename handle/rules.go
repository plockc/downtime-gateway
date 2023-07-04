package handle

import (
	"strconv"

	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func RuleChainedFactory(rule *iptables.Rule) ChainedFactory {
	return func() (ChainedFactory, Factory) {
		factory := func(ruleId string) (resource.Resource, error) {
			if ruleId != "" {
				id, err := strconv.ParseUint(ruleId, 16, 32)
				if err != nil {
					return iptables.Rule{}.RuleResource(), err
				}
				rule.Id = uint32(id)
			}
			return rule.RuleResource(), nil
		}
		return ChainChainedFactory(&rule.Chain), factory
	}
}

var Rules = Resources{
	Label: "IPTables Rules",
	ChainedFactory: func() (ChainedFactory, Factory) {
		rule := iptables.Rule{}
		return RuleChainedFactory(&rule)()
	},
	Allowed: []Allowed{LIST_ALLOWED, UPSERT_ALLOWED, GET_ALLOWED, DELETE_ALLOWED},
}
