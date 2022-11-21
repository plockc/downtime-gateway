package handle_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/handle"
	"github.com/plockc/gateway/namespace"
	"github.com/plockc/gateway/resource"
	"github.com/plockc/gateway/runner"
)

var (
	testNS = namespace.NS("test")
)

func AssertHandler(t *testing.T, handler handle.TypedHandler, parts []string, body any, expectedCode int) any {
	code, data, err := handler.Handler(parts, body)
	if err != nil {
		t.Fatal(err)
	}
	if code != expectedCode {
		t.Fatalf("expected code %d but got code %d and body: %v", expectedCode, code, data)
	}
	return data
}

func AssertHandlerFail(t *testing.T, handler handle.TypedHandler, parts []string, body any, expectedCode int) any {
	code, data, err := handler.Handler(parts, body)
	if code != expectedCode || err == nil {
		t.Fatalf(
			"expected code %d and an error, but got code %d, err: %v, and body: %v",
			expectedCode, code, err, data,
		)
	}
	return data
}

func TestMain(m *testing.M) {
	// it is the internal client outbound that can get blocked for downtime
	exitCode := func() int {
		testRunner := runner.NamespacedRunner(testNS)
		testNSLifecycle := resource.Lifecycle[string]{Resource: testNS}
		handle.NS = testNS

		testNSLifecycle.EnsureDeleted()

		// every one of these functions can error, use Do to execute, stopping if any fails
		var ignored bool
		if err := funcs.Do(
			funcs.AssignFunc(testNSLifecycle.EnsureDeleted, &ignored),
			funcs.AssignFunc(testNSLifecycle.Ensure, &ignored),
		); err != nil {
			fmt.Println(testRunner)
			fmt.Println(err)
			return 1
		}

		return m.Run()
	}()

	os.Exit(exitCode)
}
