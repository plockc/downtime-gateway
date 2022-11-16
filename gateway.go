package gateway

import (
	"strings"
)

var InternetDevice = "eth1"

const (
	DOWNTIME_CHAIN = "downtime"
)

type NetInterface string

func (ni NetInterface) IPAddrJsonCmd() []string {
	return strings.Split("ip -j -p addr show dev "+string(ni), " ")
}
