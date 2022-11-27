package iptables

type ChainCmd string

const (
	LIST         ChainCmd = "-L"
	FLUSH        ChainCmd = "-F"
	NEW          ChainCmd = "-N"
	DELETE_CHAIN ChainCmd = "-X"
)

func (ipcc ChainCmd) ChainCmd(name string) string {
	return "iptables " + string(ipcc) + " " + name
}
