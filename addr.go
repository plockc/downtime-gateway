package gateway

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type MAC [6]uint8

func (m MAC) String() string {
	return fmt.Sprintf(
		"%0.2X:%0.2X:%0.2X:%0.2X:%0.2X:%0.2X",
		m[0], m[1], m[2], m[3], m[4], m[5],
	)
}

func MACFromString(mac string) (MAC, error) {
	bs, err := hex.DecodeString(strings.ReplaceAll(mac, ":", ""))
	if err != nil {
		return MAC{}, err
	}
	return MAC{bs[0], bs[1], bs[2], bs[3], bs[4], bs[5]}, nil
}

type IPAddrsOut []IPAddrOut

type IPAddrOut struct {
	Address string `json:"address"`
	IfName  string `json:"ifname"`
}

// given an interface, can extract MAC from an IPAddrOut
// done this way so can be used in a Pipe
func MACAddrFunc(ifName string) func(IPAddrsOut) (string, error) {
	return func(ipAddrsOut IPAddrsOut) (string, error) {
		for _, i := range ipAddrsOut {
			if i.IfName == ifName {
				return i.Address, nil
			}
		}
		return "", fmt.Errorf("interface '%s' not found in output", ifName)
	}
}

// pass in the output from `ip addr -j`
func IPAddrsOutFromString(output string) (IPAddrsOut, error) {
	target := IPAddrsOut{}
	err := json.Unmarshal([]byte(output), &target)
	return target, err
}
