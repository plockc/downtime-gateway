package gateway

import (
	"strings"

	"golang.org/x/exp/slices"
)

var InternetDevice = "eth1"

const (
	DOWNTIME_CHAIN = "downtime"
)

type NetInterface string

func (ni NetInterface) IPAddrJsonCmd() []string {
	return strings.Split("ip -j -p addr show dev lan-peer", " ")
}

func EnsureDowntimeChain() error {
	chains := ListFilterChainsCmd()
	if slices.Contains(chains, DOWNTIME_CHAIN) {
		return nil
	}
	_, err := RunCmdLine(
		"iptables -C FORWARD --out-interface " + InternetDevice + " -j downtime",
	)
	return err
}

func RemoveDowntimeChainCmd() string {
	return "iptables -D FORWARD --out-interface " + InternetDevice + " -j downtime"
}
