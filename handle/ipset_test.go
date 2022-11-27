package handle_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/plockc/gateway/resource"
	"golang.org/x/exp/slices"
)

func ClearIPSets(ns resource.NS, t *testing.T, set ...string) {
	runner := testNS.Runner()
	for _, s := range set {
		if err := runner.Line("ipset destroy -exist " + s); err != nil {
			t.Fatalf("failed to clear ipsets: %v: %s", err, runner.LastOut())
		}
	}
}

func TestIPSetHandlers(t *testing.T) {
	ClearIPSets(testNS, t, "test")

	t.Run("getting set that does not exist", func(t *testing.T) {
		AssertHandlerFail(t, http.MethodGet, "/api/v1/netns/test/ipsets/test", nil, 404)
	})

	t.Run("creating set", func(t *testing.T) {
		data := AssertHandler[any](t, http.MethodPut, "/api/v1/netns/test/ipsets/test", nil, 201)
		if data != nil {
			t.Fatalf("did not expect body on create: %#v", data)
		}
	})

	t.Run("creating same set again", func(t *testing.T) {
		data := AssertHandler[any](t, http.MethodPut, "/api/v1/netns/test/ipsets/test", nil, 200)
		if data != nil {
			t.Fatalf("did not expect body on create: %#v", data)
		}
	})

	t.Run("check existence of set using GET on set name", func(t *testing.T) {
		data := AssertHandler[any](t, http.MethodGet, "/api/v1/netns/test/ipsets/test", nil, 200)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("check members of empty set", func(t *testing.T) {
		data := AssertHandler[[]string](t, http.MethodGet, "/api/v1/netns/test/ipsets/test/members", nil, 200)
		if len(*data) != 0 {
			t.Fatalf("MACs returned: %v", data)
		}
	})

	addMembersTest := func(t *testing.T) {
		t.Run("create first member", func(t *testing.T) {
			data := AssertHandler[any](
				t, http.MethodPut, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:12", nil, 201,
			)
			if data != nil {
				t.Fatalf("did not expect body: %#v", data)
			}
		})
		t.Run("add existing member again", func(t *testing.T) {
			data := AssertHandler[any](
				t, http.MethodPut, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:12", nil, 200,
			)
			if data != nil {
				t.Fatalf("did not expect body: %#v", data)
			}
		})
		t.Run("add second member", func(t *testing.T) {
			data := AssertHandler[any](
				t, http.MethodPut, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:34", nil, 201,
			)
			if data != nil {
				t.Fatalf("did not expect body: %#v", data)
			}
		})
	}
	t.Run("add members", addMembersTest)

	t.Run("check specific member", func(t *testing.T) {
		data := AssertHandler[any](
			t, http.MethodGet, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:12", nil, 200,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	mac := "12:12:12:12:12:12"
	mac2 := "12:12:12:12:12:34"

	t.Run("get members", func(t *testing.T) {
		data := AssertHandler[[]string](
			t, http.MethodGet, "/api/v1/netns/test/ipsets/test/members", nil, 200,
		)
		if !slices.Contains(*data, mac) || !slices.Contains(*data, mac2) {
			t.Fatalf("did not get expected MACs: %v", data)
		}
	})

	t.Run("remove specific existing member", func(t *testing.T) {
		data := AssertHandler[any](
			t, http.MethodDelete, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:12", nil, 204,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("remove already missing member", func(t *testing.T) {
		data := AssertHandler[any](
			t, http.MethodDelete, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:12", nil, 200,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", *data)
		}
	})

	t.Run("get remaining member", func(t *testing.T) {
		data := AssertHandler[[]string](
			t, http.MethodGet, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:34", nil, 200,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", *data)
		}
	})

	t.Run("remove second member", func(t *testing.T) {
		data := AssertHandler[any](
			t, http.MethodDelete, "/api/v1/netns/test/ipsets/test/members/12:12:12:12:12:34", nil, 204,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", *data)
		}
	})

	t.Run("get no members after deleting one by one", func(t *testing.T) {
		data := AssertHandler[[]string](
			t, http.MethodGet, "/api/v1/netns/test/ipsets/test/members", nil, 200,
		)
		if !reflect.DeepEqual(data, &[]string{}) {
			t.Fatalf("did not get empty list of MACs: %v", data)
		}
	})

	t.Run("add members", addMembersTest)

	t.Run("remove all members", func(t *testing.T) {
		data := AssertHandler[any](
			t, http.MethodDelete, "/api/v1/netns/test/ipsets/test/members", nil, 204,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})

	t.Run("get no members after deleting all at once", func(t *testing.T) {
		data := AssertHandler[[]string](
			t, http.MethodGet, "/api/v1/netns/test/ipsets/test/members", nil, 200,
		)
		if !reflect.DeepEqual(data, &[]string{}) {
			t.Fatalf("did not get empty list of MACs: %v", data)
		}
	})

	t.Run("remove all members from empty set", func(t *testing.T) {
		data := AssertHandler[any](
			t, http.MethodDelete, "/api/v1/netns/test/ipsets/test/members", nil, 200,
		)
		if data != nil {
			t.Fatalf("did not expect body: %#v", data)
		}
	})
}
