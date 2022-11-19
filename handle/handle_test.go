package handle_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/plockc/gateway"
	"github.com/plockc/gateway/handle"
)

var (
	gw = gateway.NS("test")
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
		runner := &gateway.Runner{}
		gwRunner := gateway.NamespacedRunner(gw)
		handle.Runner = gwRunner

		defer runner.Func(gw.DelCmd())

		// every one of these functions can error, use Do to execute, stopping if any fails
		if err := gateway.Do(
			runner.Func(gw.DelCmd()),
			runner.LineFunc(gw.CreateCmd()),
		); err != nil {
			fmt.Println(gwRunner)
			fmt.Println(err)
			return 1
		}

		return m.Run()
	}()

	os.Exit(exitCode)
}
