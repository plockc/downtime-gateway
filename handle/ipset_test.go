package handle_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/plockc/gateway/address"
	"github.com/plockc/gateway/handle"
	"github.com/plockc/gateway/namespace"
	"github.com/plockc/gateway/runner"
)

func ClearIPSets(ns namespace.NS, t *testing.T, set ...string) {
	gwRunner := runner.NamespacedRunner(testNS)
	for _, s := range set {
		if err := gwRunner.Line("ipset destroy -exist " + s); err != nil {
			t.Fatalf("failed to clear ipsets: %v: %s", err, gwRunner.LastOut())
		}
	}
}

func TestIPSetHandlers(t *testing.T) {
	ClearIPSets(testNS, t, "test")
	getter := handle.SetsHandlers[http.MethodGet]
	putter := handle.SetsHandlers[http.MethodPut]
	deleter := handle.SetsHandlers[http.MethodDelete]

	t.Run("getting set that does not exist", func(t *testing.T) {
		AssertHandlerFail(t, getter, []string{"test"}, nil, 404)
	})

	t.Run("creating set", func(t *testing.T) {
		data := AssertHandler(
			t, putter, []string{"test"}, nil, 201,
		)
		if data != nil {
			t.Fatalf("did not expect body on create: %#v", data)
		}
	})

	t.Run("check existence of set using GET on set name", func(t *testing.T) {
		data := AssertHandler(t, putter, []string{"test"}, nil, 200)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("check members of empty set", func(t *testing.T) {
		data := AssertHandler(t, getter, []string{"test", "members"}, nil, 200)
		if len(data.([]address.MAC)) != 0 {
			t.Fatalf("MACs returned: %v", data)
		}
	})

	addMembersTest := func(t *testing.T) {
		t.Run("create first member", func(t *testing.T) {
			data := AssertHandler(
				t, putter, []string{"test", "members", "12:12:12:12:12:12"}, nil, 201,
			)
			if data != nil {
				t.Fatalf("did not expect body: %#v", data)
			}
		})
		t.Run("add existing member again", func(t *testing.T) {
			data := AssertHandler(
				t, putter, []string{"test", "members", "12:12:12:12:12:12"}, nil, 200,
			)
			if data != nil {
				t.Fatalf("did not expect body: %#v", data)
			}
		})
		AssertHandler(
			t, putter, []string{"test", "members", "12:12:12:12:12:34"}, nil, 201,
		)
	}
	t.Run("add members", addMembersTest)

	t.Run("check specific member", func(t *testing.T) {
		data := AssertHandler(
			t, getter, []string{"test", "members", "12:12:12:12:12:12"}, nil, 200,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	mac, _ := address.MACFromString("12:12:12:12:12:12")
	mac2, _ := address.MACFromString("12:12:12:12:12:34")

	t.Run("get members", func(t *testing.T) {
		data := AssertHandler(
			t, getter, []string{"test", "members"}, nil, 200)
		if !reflect.DeepEqual(data.([]address.MAC), []address.MAC{mac, mac2}) {
			t.Fatalf("did not get expected MACs: %v", data)
		}
	})

	t.Run("remove specific existing member", func(t *testing.T) {
		data := AssertHandler(
			t, deleter, []string{"test", "members", "12:12:12:12:12:12"}, nil, 204,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("remove already missing member", func(t *testing.T) {
		data := AssertHandler(
			t, deleter, []string{"test", "members", "12:12:12:12:12:12"}, nil, 200,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("get remaining member", func(t *testing.T) {
		data := AssertHandler(
			t, getter, []string{"test", "members"}, nil, 200)
		if !reflect.DeepEqual(data.([]address.MAC), []address.MAC{mac2}) {
			t.Fatalf("did not get expected MAC: %v", data)
		}
	})

	t.Run("remove second member", func(t *testing.T) {
		data := AssertHandler(
			t, deleter, []string{"test", "members", "12:12:12:12:12:34"}, nil, 204,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("get no members after deleting one by one", func(t *testing.T) {
		data := AssertHandler(
			t, getter, []string{"test", "members"}, nil, 200)
		if !reflect.DeepEqual(data.([]address.MAC), []address.MAC{}) {
			t.Fatalf("did not get empty list of MACs: %v", data)
		}
	})

	t.Run("add members", addMembersTest)

	t.Run("remove all members", func(t *testing.T) {
		data := AssertHandler(
			t, deleter, []string{"test", "members"}, nil, 204,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("get no members after deleting all at once", func(t *testing.T) {
		data := AssertHandler(
			t, getter, []string{"test", "members"}, nil, 200)
		if !reflect.DeepEqual(data.([]address.MAC), []address.MAC{}) {
			t.Fatalf("did not get empty list of MACs: %v", data)
		}
	})

	t.Run("remove all members from empty set", func(t *testing.T) {
		data := AssertHandler(
			t, deleter, []string{"test", "members"}, nil, 200,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})
}
