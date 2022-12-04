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
	var table = iptables.NewTable(ns, "filter")
	chain := iptables.NewChain(table, "")
	if err := funcs.Do(
		resource.NewLifecycle(
			chain.ChainResource(),
		).Clear,
		gwRunner.BatchLinesFunc(iptables.FLUSH.ChainCmd("FORWARD")),
		gwRunner.BatchLinesFunc(iptables.FLUSH.ChainCmd("OUTPUT")),
		gwRunner.BatchLinesFunc(iptables.FLUSH.ChainCmd("INPUT")),
	); err != nil {
		t.Error(gwRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		t.Error(iptables)
		t.Fatalf("failed to clear IPTables: %v", err)
	}
}

func ClearIPSets(ns resource.NS, t *testing.T, set ...string) {
	gwRunner := gw.Runner()
	for _, s := range set {
		if err := gwRunner.RunLine("ipset destroy -exist " + s); err != nil {
			t.Fatalf("failed to clear ipsets: %v", err)
		}
	}
}

// first test if we can route
func TestAllow(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gw.Runner()
	clientRunner := client.Runner()
	if err := clientRunner.RunLine(PingCmd(serverIP)); err != nil {
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
		gwRunner.BatchLinesFunc(iptables.APPEND.FilterRule("FORWARD", "-s 192.168.100.20", "DROP")),
		clientRunner.BatchLinesFunc(PingCmd(serverIP)),
	); err == nil {
		fmt.Println(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		fmt.Println(iptables)
		t.Fatal("did not block")
	}
}

// test if we can block a set
func TestSetBlock(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gw.Runner()
	clientRunner := client.Runner()
	var ipSet iptables.IPSet = iptables.NewIPSet(gw, "test")
	ClearIPSets(gw, t, ipSet.Name)
	ipSetRes := ipSet.IPSetResource()
	ipSetMemberRes := iptables.NewMember(ipSet, clientMAC).MemberResource()
	table := iptables.FilterTable(gw)
	chain := iptables.NewChain(table, iptables.DOWNTIME_CHAIN)
	chainRes := chain.ChainResource()
	jumpToChainRule := iptables.NewRule(iptables.NewChain(table, "FORWARD"))
	jumpToChainRule.Target = chain.Name
	rule := iptables.NewRule(chain)
	rule.Target = "DROP"
	rule.MatchSetSrc = ipSet.Name
	ruleRes := rule.RuleResource()
	defer resource.NewLifecycle(ruleRes).EnsureDeleted()
	defer resource.NewLifecycle(jumpToChainRule.RuleResource()).EnsureDeleted()
	defer resource.NewLifecycle(chainRes).EnsureDeleted()
	defer resource.NewLifecycle(ipSetMemberRes).EnsureDeleted()
	defer resource.NewLifecycle(ipSetRes).EnsureDeleted()
	if err := funcs.Do(
		ipSetRes.Create,
		ipSetMemberRes.Create,
		chainRes.Create,
		jumpToChainRule.RuleResource().Create,
		ruleRes.Create,
		funcs.ExpectFailFunc("ping server", clientRunner.BatchLinesFunc(PingCmd(serverIP))),
	); err != nil {
		t.Error(gwRunner)
		fmt.Println(clientRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		t.Error(iptables)
		_, ipsets, _ := exec.ExecLine(gwRunner.WrapCmdLine("ipset list"))
		t.Error(ipsets)
		_, ipaddrs, _ := exec.ExecLine(gwRunner.WrapCmdLine("ip a"))
		t.Error(ipaddrs)
		t.Fatal(err)
	}
}

// test if we can allow a set
func TestSetAllow(t *testing.T) {
	ClearIPTables(gw, t)
	gwRunner := gw.Runner()
	clientRunner := client.Runner()
	ipSet := iptables.NewIPSet(gw, "test")
	ClearIPSets(gw, t, ipSet.Name)
	ipSetRes := ipSet.IPSetResource()
	ipSetMemberRes := iptables.NewMember(ipSet, clientMAC).MemberResource()
	t.Log("Member:", ipSetMemberRes)
	table := iptables.FilterTable(gw)
	chain := iptables.NewChain(table, iptables.DOWNTIME_CHAIN)
	chainRes := chain.ChainResource()
	jumpToChainRule := iptables.NewRule(iptables.NewChain(table, "FORWARD"))
	jumpToChainRule.Target = chain.Name
	rule := iptables.NewRule(chain)
	rule.Target = "RETURN"
	rule.MatchSetSrc = ipSet.Name
	ruleRes := rule.RuleResource()
	defer resource.NewLifecycle(ruleRes).EnsureDeleted()
	defer resource.NewLifecycle(jumpToChainRule.RuleResource()).EnsureDeleted()
	defer resource.NewLifecycle(chainRes).EnsureDeleted()
	defer resource.NewLifecycle(ipSetMemberRes).EnsureDeleted()
	defer resource.NewLifecycle(ipSetRes).EnsureDeleted()
	if err := funcs.Do(
		ipSetRes.Create,
		ipSetMemberRes.Create,
		chainRes.Create,
		jumpToChainRule.RuleResource().Create,
		ruleRes.Create,
		clientRunner.BatchLinesFunc(PingCmd(serverIP)),
	); err != nil {
		t.Error(gwRunner)
		t.Log(clientRunner)
		_, iptables, _ := exec.ExecLine(gwRunner.WrapCmdLine("iptables -L -v"))
		t.Error(iptables)
		_, ipsets, _ := exec.ExecLine(gwRunner.WrapCmdLine("ipset list"))
		t.Error(ipsets)
		_, ipaddrs, _ := exec.ExecLine(gwRunner.WrapCmdLine("ip a"))
		t.Error(ipaddrs)
		t.Fatal(err)
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
			lc := resource.Lifecycle{Resource: ns.NSResource()}
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
			gwRunner.BatchLinesFunc("ip link add wan type veth peer name wan-peer"),
			gwRunner.BatchLinesFunc("ip link set wan-peer netns "+server.NSName()),
			gwRunner.BatchLinesFunc("ip addr add "+gwWanIP.String()+" dev wan"),
			gwRunner.BatchLinesFunc("ip link set dev wan up"),

			serverRunner.BatchLinesFunc("ip addr add "+serverIP.String()+" dev wan-peer"),
			serverRunner.BatchLinesFunc("ip link set dev wan-peer up"),
			serverRunner.BatchLinesFunc("ip route add default via "+gwWanIP.IP.String()+" dev wan-peer"),

			serverRunner.BatchLinesFunc("ip addr"),

			// connect and configure lan interfaces on the 192.168.100/24 network
			// internal client routes through gatway at 192.168.100.1
			gwRunner.BatchLinesFunc("ip link add lan type veth peer name lan-peer"),
			gwRunner.BatchLinesFunc("ip link set lan-peer netns "+client.NSName()),
			gwRunner.BatchLinesFunc("ip addr add 192.168.100.1/24 dev lan"),
			gwRunner.BatchLinesFunc("ip link set dev lan up"),

			clientRunner.BatchLinesFunc("ip addr add 192.168.100.20/24 dev lan-peer"),
			clientRunner.BatchLinesFunc("ip link set dev lan-peer up"),
			clientRunner.BatchLinesFunc("ip route add default via 192.168.100.1 dev lan-peer"),

			clientRunner.BatchFunc(address.NetInterface("lan-peer").IPAddrJsonCmd()),

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
