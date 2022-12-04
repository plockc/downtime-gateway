package handle_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func TestChainHandlers(t *testing.T) {
	table := iptables.NewTable(testNS, "filter")
	chain := iptables.NewChain(table, "testChain")

	var ignored bool
	if err := funcs.Do(
		funcs.AssignFunc(resource.NewLifecycle(table.TableResource()).Ensure, &ignored),
		chain.ChainResource().Clear,
	); err != nil {
		t.Fatal(err)
	}
	defer resource.NewLifecycle(chain.ChainResource()).EnsureDeleted()
	defer resource.NewLifecycle(table.TableResource()).EnsureDeleted()

	// at this point should be no chains

	t.Run("creating chain", func(t *testing.T) {
		data := AssertHandler[any](t, http.MethodPut, "/api/v1/netns/test/iptables/filter/chains/testChain", nil, 201)
		if data != nil {
			t.Fatalf("did not expect body on create: %#v", data)
		}
	})

	t.Run("check existence of chain using GET on chain name", func(t *testing.T) {
		data := AssertHandler[iptables.Chain](t, http.MethodGet, "/api/v1/netns/test/iptables/filter/chains/testChain", nil, 200)
		if data != nil {
			t.Fatalf("did not expect body: %#v", *data)
		}
	})

	t.Run("check existence of chain using GET for list of chains", func(t *testing.T) {
		data := AssertHandler[[]string](t, http.MethodGet, "/api/v1/netns/test/iptables/filter/chains", nil, 200)
		if data == nil || !reflect.DeepEqual(*data, []string{"INPUT", "FORWARD", "OUTPUT", "testChain"}) {
			t.Fatalf("expected testChain, got %v", *data)
		}
	})

	t.Run("remove chain", func(t *testing.T) {
		data := AssertHandler[[]string](t, http.MethodDelete, "/api/v1/netns/test/iptables/filter/chains/testChain", nil, 204)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})
}
