package iptables_test

import (
	"reflect"
	"testing"

	"github.com/plockc/gateway/address"
	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/iptables"
)

func TestRuleResource(t *testing.T) {
	table := iptables.FilterTable(testNS)
	chain := iptables.NewChain(table, "tchain")
	chainRes := iptables.NewChainResource(chain)
	ruleLister := iptables.NewRule(chain).RuleResource()
	ipSet := iptables.NewIPSet(testNS, "testSet")
	mac, err := address.MACFromString("12:12:12:12:12:12")
	if err != nil {
		t.Fatal(err)
	}
	ipSetMember := iptables.NewMember(ipSet, mac)

	// create the chain and set for testing
	err = funcs.Do(
		chainRes.Create,
		ipSet.IPSetResource().Create,
		ipSetMember.MemberResource().Create,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer chainRes.Delete()
	defer ipSet.IPSetResource().Delete()

	// ensure list starts empty
	err = ruleLister.Clear()
	if err != nil {
		t.Fatal(err)
	}
	rules, err := ruleLister.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected no rules: %v", rules)
	}

	rule := iptables.NewRule(chain)
	rule.MatchSetSrc = ipSet.Name
	rule.Comment = "testRule"
	ruleRes := rule.RuleResource()

	// create chain then test it shows in list
	err = ruleRes.Create()
	if err != nil {
		t.Fatal(err)
	}
	rules, err = ruleLister.List()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rules, []string{ruleRes.Id()}) {
		t.Fatalf("expected %v, got: %v", rule, rules)
	}

	// Delete the chain then check List is empty
	err = ruleRes.Delete()
	if err != nil {
		t.Fatal(err)
	}
	rules, err = ruleLister.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected no rules after delete: %v", rules)
	}

	rule2 := iptables.NewRule(chain)
	rule2.MatchSetSrc = ipSet.Name
	rule2.Comment = "testRule2"
	rule2Res := rule.RuleResource()

	// create two rules, test both in list
	err = ruleRes.Create()
	if err != nil {
		t.Fatal(err)
	}
	err = rule2Res.Create()
	if err != nil {
		t.Fatal(err)
	}
	rules, err = ruleLister.List()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rules, []string{ruleRes.Id(), rule2Res.Id()}) {
		t.Fatalf("expected both rules in list: %v", rules)
	}

	// clear chain then verify empty
	err = ruleLister.Clear()
	if err != nil {
		t.Fatal(err)
	}
	rules, err = ruleLister.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected no rules after Clear(): %v", rules)
	}
}
