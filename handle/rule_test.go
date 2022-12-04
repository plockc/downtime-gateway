package handle_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/plockc/gateway/address"
	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func TestRuleHandlers(t *testing.T) {
	ipSet := iptables.NewIPSet(testNS, "testSet")
	table := iptables.NewTable(testNS, "filter")
	chain := iptables.NewChain(table, "testChain")
	rule := iptables.NewRule(chain)
	rule.Comment = "this is a test rule"
	rule.MatchSetSrc = ipSet.Name
	rule.Target = "DROP"

	var ignored bool
	var mac address.MAC
	if err := funcs.Do(
		funcs.AssignFunc(resource.NewLifecycle(ipSet.IPSetResource()).Ensure, &ignored),
		funcs.AssignFunc(func() (address.MAC, error) {
			return address.MACFromString("12:12:12:12:12:12")
		}, &mac),
		funcs.AssignFunc(resource.NewLifecycle(table.TableResource()).Ensure, &ignored),
		funcs.AssignFunc(resource.NewLifecycle(chain.ChainResource()).Ensure, &ignored),
		rule.RuleResource().Clear,
	); err != nil {
		t.Fatal(err)
	}
	defer resource.NewLifecycle(rule.RuleResource()).EnsureDeleted()
	defer resource.NewLifecycle(chain.ChainResource()).EnsureDeleted()
	defer resource.NewLifecycle(table.TableResource()).EnsureDeleted()

	member := iptables.NewMember(ipSet, mac)
	defer resource.NewLifecycle(member.MemberResource()).EnsureDeleted()
	defer resource.NewLifecycle(ipSet.IPSetResource()).EnsureDeleted()

	// at this point should be no rules

	var createdRuleId string
	t.Run("creating rule", func(t *testing.T) {
		rulesPath := "/api/v1/netns/test/iptables/filter/chains/testChain/rules"
		body, headers := AssertHandlerGetHeaders[any](
			t, http.MethodPut, rulesPath, rule, 201,
		)
		if body != nil {
			t.Fatalf("got unexpected body: %v", *body)
		}
		location := headers.Get("Location")
		if !strings.HasPrefix(location, rulesPath) {
			t.Fatalf("unexpected location, got '%s'", location)
		}
		createdRuleId = strings.TrimLeft(location, rulesPath+"/")
	})

	t.Run("check existence of rule using GET on rule id", func(t *testing.T) {
		data := AssertHandler[iptables.Rule](
			t, http.MethodGet, "/api/v1/netns/test/iptables/filter/chains/testChain/rules/"+createdRuleId, nil, 200,
		)
		if data == nil {
			t.Fatalf("missing body: %#v", data)
		}
		if data.Target != "DROP" {
			t.Fatalf("wrong target, had: %s", data.Target)
		}
		if data.MatchSetSrc != "testSet" {
			t.Fatalf("wrong set name, had: %s", data.MatchSetSrc)
		}
		if data.Comment != rule.Comment {
			t.Fatalf("wrong comment, had: %s", data.Comment)
		}
	})

	t.Run("check existence of rule using GET for list of rules", func(t *testing.T) {
		data := AssertHandler[[]string](t, http.MethodGet, "/api/v1/netns/test/iptables/filter/chains/testChain/rules", nil, 200)
		if len(*data) != 1 {
			t.Fatalf("expected rule, got %v", *data)
		}
		ruleId := (*data)[0]
		if ruleId != createdRuleId {
			t.Fatalf("GET returned Id '%s' instead of expected '%s'", ruleId, createdRuleId)
		}
	})

	t.Run("remove rule", func(t *testing.T) {
		data := AssertHandler[[]string](t, http.MethodDelete, "/api/v1/netns/test/iptables/filter/chains/testChain/rules/"+createdRuleId, nil, 204)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})
}
