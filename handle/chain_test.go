package handle_test

import (
	"reflect"
	"testing"

	"github.com/plockc/gateway/iptables"
)

func TestChainHandlers(t *testing.T) {
	table := iptables.FilterTable(testNS)
	chain := iptables.NewChain(table, "tchain")
	chainNames := func(cs ...iptables.Chain) []string {
		chains := []string{"INPUT", "FORWARD", "OUTPUT"}
		for _, c := range cs {
			chains = append(chains, c.Name)
		}
		return chains
	}
	chainRes := iptables.NewChainResource(chain)

	// ensure list starts empty
	err := chainRes.Clear()
	if err != nil {
		t.Fatal(err)
	}
	chains, err := chainRes.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(chains) != 3 {
		t.Fatalf("expected standard 3 chains: %v", chains)
	}

	// create chain then test it shows in list
	err = chainRes.Create()
	if err != nil {
		t.Fatal(err)
	}
	chains, err = chainRes.List()
	if err != nil {
		t.Fatal(err)
	}
	expectedChains := chainNames(chain)
	if !reflect.DeepEqual(chains, expectedChains) {
		t.Fatalf("expected %v, got: %v", expectedChains, chains)
	}

	// Delete the chain then check List is empty
	err = chainRes.Delete()
	if err != nil {
		t.Fatal(err)
	}
	chains, err = chainRes.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(chains) != 3 {
		t.Fatalf("expected standard 3 chains after delete: %v", chains)
	}

	// creeate two chain, test both in list
	chain2 := iptables.NewChain(table, "tchain2")
	err = chainRes.Create()
	if err != nil {
		t.Fatal(err)
	}
	err = iptables.NewChainResource(chain2).Create()
	if err != nil {
		t.Fatal(err)
	}
	chains, err = chainRes.List()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(chains, chainNames(chain, chain2)) {
		t.Fatalf("expected both chains in list: %v", chains)
	}

	// clear chain then verify empty
	err = chainRes.Clear()
	if err != nil {
		t.Fatal(err)
	}
	chains, err = chainRes.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(chains) != 3 {
		t.Fatalf("expected standard 3 chains after Clear(): %v", chains)
	}
}
