package gateway_test

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/plockc/gateway"
)

func init() {
	gateway.InternetDevice = "wan"
}

var (
	server = gateway.NS("public-server")
	client = gateway.NS("internal-client")
	gw     = gateway.NS("gateway")

	serverIP = net.IPNet{IP: net.ParseIP("44.44.44.44"), Mask: net.CIDRMask(16, 32)}
	gwWanIP  = net.IPNet{IP: net.ParseIP("44.44.55.55"), Mask: net.CIDRMask(16, 32)}

	clientMAC gateway.MAC
)

func PingCmd(target net.IPNet) string {
	return "ping -c 1 -W 1 " + target.IP.String()
}

func ClearIPTables(ns gateway.NS, t *testing.T) {
	gwRunner := gateway.NamespacedRunner(gw)
	if err := gateway.Do(
		gateway.RemoveChainCmdFunc(gwRunner, gateway.DOWNTIME_CHAIN),
		gwRunner.LineFunc(gateway.FLUSH.ChainCmd("FORWARD")),
		gwRunner.LineFunc(gateway.FLUSH.ChainCmd("OUTPUT")),
		gwRunner.LineFunc(gateway.FLUSH.ChainCmd("INPUT")),
	); err != nil {
		fmt.Println(gwRunner)
		_, iptables, _ := gateway.ExecLine(gw.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("failed to clear IPTables")
	}
}

// first test if we can route
func TestAllow(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gateway.NamespacedRunner(gw)
	clientRunner := gateway.NamespacedRunner(client)
	if err := clientRunner.Line(PingCmd(serverIP)); err != nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := gateway.ExecLine(gw.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal(err)
	}
}

// test if we can block
func TestRuleBlock(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gateway.NamespacedRunner(gw)
	clientRunner := gateway.NamespacedRunner(client)
	if err := gateway.Do(
		gwRunner.LineFunc(gateway.APPEND.FilterRule("FORWARD", "-s 192.168.100.20", "DROP")),
		clientRunner.LineFunc(PingCmd(serverIP)),
	); err == nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := gateway.ExecLine(gw.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("did not block")
	}
}

// test if we can block a set
func TestBlockSet(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gateway.NamespacedRunner(gw)
	clientRunner := gateway.NamespacedRunner(client)
	ipSet := gateway.IPSet{Name: "test", Members: []gateway.MAC{clientMAC}}
	if err := gateway.Do(
		gwRunner.LineFunc(ipSet.SyncCmdLines()...),
		gateway.EnsureChainFunc(gwRunner, gateway.DOWNTIME_CHAIN),
		gwRunner.LineFunc(gateway.APPEND.FilterRule(
			gateway.DOWNTIME_CHAIN, ipSet.Match(), gateway.DROP),
		),
		gateway.ExpectFailFunc("ping server", clientRunner.LineFunc(PingCmd(serverIP))),
	); err != nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := gateway.ExecLine(gw.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("failed to block")
	}
}

// test if we can allow a set
func TestAllowSet(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gateway.NamespacedRunner(gw)
	clientRunner := gateway.NamespacedRunner(client)
	ipSet := gateway.IPSet{Name: "test", Members: []gateway.MAC{clientMAC}}
	if err := gateway.Do(
		gwRunner.LineFunc(ipSet.SyncCmdLines()...),
		gateway.EnsureChainFunc(gwRunner, gateway.DOWNTIME_CHAIN),
		gwRunner.LineFunc(gateway.APPEND.FilterRule(
			gateway.DOWNTIME_CHAIN, ipSet.Match(), gateway.RETURN),
		),
		clientRunner.LineFunc(PingCmd(serverIP)),
	); err != nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := gateway.ExecLine(gw.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("blocked")
	}
}

func TestMain(m *testing.M) {
	// it is the internal client outbound that can get blocked for downtime
	exitCode := func() int {
		runner := &gateway.Runner{}
		gwRunner := gateway.NamespacedRunner(gw)
		clientRunner := gateway.NamespacedRunner(client)
		serverRunner := gateway.NamespacedRunner(server)

		// these are deferred to end of this inner function
		for _, h := range []gateway.NS{server, client, gw} {
			defer runner.Func(h.DelCmd())
		}

		// every one of these functions can error, use Do to execute, stopping if any fails
		if err := gateway.Do(
			runner.Func(server.DelCmd()),
			runner.Func(client.DelCmd()),
			runner.Func(gw.DelCmd()),
			runner.LineFunc(server.CreateCmd()),
			runner.LineFunc(client.CreateCmd()),
			runner.LineFunc(gw.CreateCmd()),

			// connect and configure wan interfaces in the 44.0.0.0/8 network
			// the internal server normally has many routes but for this test
			// just route to the test gateway for return traffic
			gwRunner.LineFunc("ip link add wan type veth peer name wan-peer"),
			gwRunner.LineFunc("ip link set wan-peer netns "+string(server)),
			gwRunner.LineFunc("ip addr add "+gwWanIP.String()+" dev wan"),
			gwRunner.LineFunc("ip link set dev wan up"),

			serverRunner.LineFunc("ip addr add "+serverIP.String()+" dev wan-peer"),
			serverRunner.LineFunc("ip link set dev wan-peer up"),
			serverRunner.LineFunc("ip route add default via "+gwWanIP.IP.String()+" dev wan-peer"),

			serverRunner.LineFunc("ip addr"),

			// connect and configure lan interfaces on the 192.168.100/24 network
			// internal client routes through gatway at 192.168.100.1
			gwRunner.LineFunc("ip link add lan type veth peer name lan-peer"),
			gwRunner.LineFunc("ip link set lan-peer netns "+string(client)),
			gwRunner.LineFunc("ip addr add 192.168.100.1/24 dev lan"),
			gwRunner.LineFunc("ip link set dev lan up"),

			clientRunner.LineFunc("ip addr add 192.168.100.20/24 dev lan-peer"),
			clientRunner.LineFunc("ip link set dev lan-peer up"),
			clientRunner.LineFunc("ip route add default via 192.168.100.1 dev lan-peer"),

			clientRunner.Func(gateway.NetInterface("lan-peer").IPAddrJsonCmd()),

			gateway.AssignFunc(clientRunner.LastOutFunc(), &clientMAC, gateway.Pipe(gateway.Pipe(
				gateway.IPAddrsOutFromString,
				gateway.MACAddrFunc("lan-peer")),
				gateway.MACFromString,
			)),
		); err != nil {
			fmt.Println(runner)
			fmt.Println(gwRunner)
			fmt.Println(clientRunner)
			fmt.Println(err)
			return 1
		}

		return m.Run()
	}()

	os.Exit(exitCode)
}
