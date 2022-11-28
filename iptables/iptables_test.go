package iptables_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/handle"
	"github.com/plockc/gateway/resource"
)

var (
	testNS = resource.NewNS("iptablestest")
)

func TestMain(m *testing.M) {
	// it is the internal client outbound that can get blocked for downtime
	exitCode := func() int {
		testRunner := testNS.Runner()
		testNSLifecycle := resource.Lifecycle{Resource: testNS.NSResource()}
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
