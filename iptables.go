package gateway

type IPRuleCmd string
type IPChainCmd string

const (
	APPEND IPRuleCmd = "-A"
	CHECK  IPRuleCmd = "-C"
	DELETE IPRuleCmd = "-D"

	LIST         IPChainCmd = "-L"
	FLUSH        IPChainCmd = "-F"
	NEW          IPChainCmd = "-N"
	DELETE_CHAIN IPChainCmd = "-X"
)

func (iptc IPRuleCmd) FilterRule(chain, match, target string) string {
	return "iptables " + string(iptc) + " " + chain + " " + match + " -j " + target
}

func ListFilterChainsCmd() []string {
	return []string{
		"bash", "-c", `iptables-save \
		| sed -n '/^*filter/,/^[^:]/{/^:/!d;s/:\(\w*\) .*/\1/;p}'`,
	}
}
