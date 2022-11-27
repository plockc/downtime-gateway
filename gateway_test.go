package gateway_test

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/plockc/gateway/address"
	"github.com/plockc/gateway/exec"
	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
)

func init() {
	iptables.InternetDevice = "wan"
}

var (
	server = resource.NewNS("public-server")
	client = resource.NewNS("internal-client")
	gw     = resource.NewNS("gateway")

	serverIP = net.IPNet{IP: net.ParseIP("44.44.44.44"), Mask: net.CIDRMask(16, 32)}
	gwWanIP  = net.IPNet{IP: net.ParseIP("44.44.55.55"), Mask: net.CIDRMask(16, 32)}

	clientMAC address.MAC
)

func PingCmd(target net.IPNet) string {
	return "ping -c 1 -W 1 " + target.IP.String()
}

func ClearIPTables(ns resource.NS, t *testing.T) {
	gwRunner := gw.Runner()
	if err := funcs.Do(
		iptables.RemoveChainCmdFunc(gwRunner, iptables.DOWNTIME_CHAIN),
		gwRunner.LineFunc(iptables.FLUSH.ChainCmd("FORWARD")),
		gwRunner.LineFunc(iptables.FLUSH.ChainCmd("OUTPUT")),
		gwRunner.LineFunc(iptables.FLUSH.ChainCmd("INPUT")),
	); err != nil {
		fmt.Println(gwRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatalf("failed to clear IPTables: %v", err)
	}
}

func ClearIPSets(ns resource.NS, t *testing.T, set ...string) {
	gwRunner := gw.Runner()
	for _, s := range set {
		if err := gwRunner.Line("ipset destroy -exist " + s); err != nil {
			t.Fatalf("failed to clear ipsets: %v", err)
		}
	}
}

// first test if we can route
func TestAllow(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gw.Runner()
	clientRunner := client.Runner()
	if err := clientRunner.Line(PingCmd(serverIP)); err != nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal(err)
	}
}

// test if we can block
func TestRuleBlock(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gw.Runner()
	clientRunner := client.Runner()
	if err := funcs.Do(
		gwRunner.LineFunc(iptables.APPEND.FilterRule("FORWARD", "-s 192.168.100.20", "DROP")),
		clientRunner.LineFunc(PingCmd(serverIP)),
	); err == nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("did not block")
	}
}

// test if we can block a set
func TestBlockSet(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gw.Runner()
	clientRunner := client.Runner()
	ipSet := iptables.NewIPSet(gw, "test")
	ipSetRes := ipSet.Resource()
	ipSetMemberRes := iptables.NewMember(ipSet, clientMAC).Resource()
	if err := funcs.Do(
		ipSetRes.Create,
		ipSetMemberRes.Create,
		iptables.EnsureChainFunc(gwRunner, iptables.DOWNTIME_CHAIN),
		gwRunner.LineFunc(iptables.APPEND.FilterRule(
			iptables.DOWNTIME_CHAIN, ipSet.Match(), iptables.DROP),
		),
		funcs.ExpectFailFunc("ping server", clientRunner.LineFunc(PingCmd(serverIP))),
	); err != nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("failed to block")
	}
}

// test if we can allow a set
func TestAllowSet(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gw.Runner()
	clientRunner := client.Runner()
	ipSet := iptables.NewIPSet(gw, "test")
	ipSetLC := resource.Lifecycle{Resource: ipSet.Resource()}
	ipSetMember := iptables.NewMember(ipSet, clientMAC)
	ipSetMemberLC := resource.Lifecycle{Resource: ipSetMember.Resource()}
	var ignored bool
	if err := funcs.Do(
		funcs.AssignFunc(ipSetLC.Ensure, &ignored),
		funcs.AssignFunc(ipSetMemberLC.Ensure, &ignored),
		iptables.EnsureChainFunc(gwRunner, iptables.DOWNTIME_CHAIN),
		gwRunner.LineFunc(iptables.APPEND.FilterRule(
			iptables.DOWNTIME_CHAIN, ipSet.Match(), iptables.RETURN),
		),
		clientRunner.LineFunc(PingCmd(serverIP)),
	); err != nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("blocked")
	}
}

func TestMain(m *testing.M) {
	// it is the internal client outbound that can get blocked for downtime
	exitCode := func() int {
		run := resource.NewNS("").Runner()
		gwRunner := gw.Runner()
		clientRunner := client.Runner()
		serverRunner := server.Runner()

		var ignored bool
		for _, ns := range []resource.NS{server, client, gw} {
			lc := resource.Lifecycle{Resource: ns.Resource()}
			if err := funcs.Do(
				funcs.AssignFunc(lc.EnsureDeleted, &ignored),
				funcs.AssignFunc(lc.Ensure, &ignored),
			); err != nil {
				fmt.Println(err)
				return 1
			}
			// these cleanups are deferred to end of this inner function
			defer lc.EnsureDeleted()
		}

		// every one of these functions can error, use Do to execute, stopping if any fails
		if err := funcs.Do(
			// connect and configure wan interfaces in the 44.0.0.0/8 network
			// the internal server normally has many routes but for this test
			// just route to the test gateway for return traffic
			gwRunner.LineFunc("ip link add wan type veth peer name wan-peer"),
			gwRunner.LineFunc("ip link set wan-peer netns "+server.NSName()),
			gwRunner.LineFunc("ip addr add "+gwWanIP.String()+" dev wan"),
			gwRunner.LineFunc("ip link set dev wan up"),

			serverRunner.LineFunc("ip addr add "+serverIP.String()+" dev wan-peer"),
			serverRunner.LineFunc("ip link set dev wan-peer up"),
			serverRunner.LineFunc("ip route add default via "+gwWanIP.IP.String()+" dev wan-peer"),

			serverRunner.LineFunc("ip addr"),

			// connect and configure lan interfaces on the 192.168.100/24 network
			// internal client routes through gatway at 192.168.100.1
			gwRunner.LineFunc("ip link add lan type veth peer name lan-peer"),
			gwRunner.LineFunc("ip link set lan-peer netns "+client.NSName()),
			gwRunner.LineFunc("ip addr add 192.168.100.1/24 dev lan"),
			gwRunner.LineFunc("ip link set dev lan up"),

			clientRunner.LineFunc("ip addr add 192.168.100.20/24 dev lan-peer"),
			clientRunner.LineFunc("ip link set dev lan-peer up"),
			clientRunner.LineFunc("ip route add default via 192.168.100.1 dev lan-peer"),

			clientRunner.Func(address.NetInterface("lan-peer").IPAddrJsonCmd()),

			funcs.PipeFunc(clientRunner.LastOutFunc(), &clientMAC, funcs.Pipe(funcs.Pipe(
				address.IPAddrsOutFromString,
				address.MACAddrFunc("lan-peer")),
				address.MACFromString,
			)),
		); err != nil {
			fmt.Println(run)
			fmt.Println(gwRunner)
			fmt.Println(clientRunner)
			fmt.Println(err)
			return 1
		}

		return m.Run()
	}()

	os.Exit(exitCode)
}
