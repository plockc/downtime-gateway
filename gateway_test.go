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

// first test if we can route
func TestAllow(t *testing.T) {
	if _, err := gateway.RunCmdLine(client.PingCmdLine(serverIP)); err != nil {
		t.Fatal(err)
	}
}

// test if we can block
func TestRuleBlock(t *testing.T) {
	if _, err := gateway.RunCmdLines(
		client.WrapCmdLine("iptables -A FORWARD -s 192.168.100.20 -j DROP"),
		client.PingCmdLine(serverIP),
	); err == nil {
		t.Fatal("did not block")
	}
}

// test if we can block a set
func TestBlockSet(t *testing.T) {
	ipSet := gateway.IPSet{Name: "test", Members: []gateway.MAC{clientMAC}}
	cmds := gateway.Cmds{}
	cmds.AddCmdLine(gw.WrapCmdLines(ipSet.CreateCmdLines())...)
	cmds.AddCmdLine(gw.WrapCmdLine(ipSet.BlockInternetCmdLine()))
	cmds.AddCmdLine(client.PingCmdLine(serverIP))
	if outs, err := gateway.Run(cmds...); err == nil {
		fmt.Println(gateway.Multiline(cmds.Debug(outs)))
		t.Fatal("did not block")
	}
}

// test if we can allow a set
func TestAllowSet(t *testing.T) {
	ipSet := gateway.IPSet{Name: "test", Members: []gateway.MAC{clientMAC}}
	cmds := gateway.Cmds{}
	cmds.AddCmdLine(gw.WrapCmdLines(ipSet.CreateCmdLines())...)
	cmds.AddCmdLine(gw.WrapCmdLine("ipset list"))
	cmds.AddCmdLine(gw.WrapCmdLine("ip addr"))
	// FIX
	//cmds.AddCmdLine(gw.WrapCmdLine(EnsureDowntimeChain()))
	//	cmds.AddCmdLine(gw.WrapCmdLine(ipSet.ReturnInternetCmdLine()))
	cmds.AddCmdLine(client.PingCmdLine(serverIP))
	if outs, err := gateway.Run(cmds...); err != nil {
		fmt.Println(gateway.Multiline(cmds.Debug(outs)))
		t.Fatal("blocked")
	}
}

func TestMain(m *testing.M) {
	// it is the internal client outbound that can get blocked for downtime
	cmds := &gateway.Cmds{}

	exitCode := func() int {
		for _, h := range []gateway.NS{server, client, gw} {
			cmds.Add(h.DelCmd())
			cmds.AddCmdLine("ip netns add " + string(h))
			defer gateway.Run(h.DelCmd())
		}

		// connect and configure wan interfaces in the 44.0.0.0/8 network
		// the internal server normally has many routes but for this test
		// just route to the test gateway for return traffic
		cmds.AddCmdLine(gw.WrapCmdLine("ip link add wan type veth peer name wan-peer"))
		cmds.AddCmdLine(gw.WrapCmdLine("ip link set wan-peer netns " + string(server)))
		cmds.AddCmdLine(gw.WrapCmdLine("ip addr add " + gwWanIP.String() + " dev wan"))
		cmds.AddCmdLine(gw.WrapCmdLine("ip link set dev wan up"))

		cmds.AddCmdLine(server.WrapCmdLine("ip addr add " + serverIP.String() + " dev wan-peer"))
		cmds.AddCmdLine(server.WrapCmdLine("ip link set dev wan-peer up"))
		cmds.AddCmdLine(server.WrapCmdLine("ip route add default via " + gwWanIP.IP.String() + " dev wan-peer"))

		cmds.AddCmdLine(server.WrapCmdLine("ip addr"))

		// connect and configure lan interfaces on the 192.168.100/24 network
		// internal client routes through gatway at 192.168.100.1
		cmds.AddCmdLine(gw.WrapCmdLine("ip link add lan type veth peer name lan-peer"))
		cmds.AddCmdLine(gw.WrapCmdLine("ip link set lan-peer netns " + string(client)))
		cmds.AddCmdLine(gw.WrapCmdLine("ip addr add 192.168.100.1/24 dev lan"))
		cmds.AddCmdLine(gw.WrapCmdLine("ip link set dev lan up"))

		cmds.AddCmdLine(client.WrapCmdLine("ip addr add 192.168.100.20/24 dev lan-peer"))
		cmds.AddCmdLine(client.WrapCmdLine("ip link set dev lan-peer up"))
		cmds.AddCmdLine(client.WrapCmdLine("ip route add default via 192.168.100.1 dev lan-peer"))

		clientAddrsIdx := cmds.Add(client.WrapCmd(gateway.NetInterface("lan-peer").IPAddrJsonCmd()))[0]

		outs, err := gateway.Run(*cmds...)
		if err != nil {
			fmt.Println(err)
			fmt.Println(gateway.Multiline(cmds.Debug(outs)))
			return 1
		}

		clientMAC, err = gateway.Pipe(gateway.Pipe(
			gateway.IPAddrsOutFromString,
			gateway.MACAddrFunc("lan-peer")),
			gateway.MACFromString,
		)(outs[clientAddrsIdx])
		if err != nil {
			fmt.Println("Failed to get MAC address from interfaces: %w", err)
			return 1
		}

		return m.Run()
	}()

	os.Exit(exitCode)
}
