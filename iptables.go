package gateway

import (
	"strings"

	"golang.org/x/exp/slices"
)

type IPRuleCmd string
type IPChainCmd string

const (
	DROP   = "DROP"
	RETURN = "RETURN"

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

func EnsureIPRuleFunc(runner Runner, chain, match, target string) func() error {
	return func() error {
		err := runner.Line(CHECK.FilterRule(chain, match, target))
		if err != nil {
			return err
		}
		return runner.Line(APPEND.FilterRule(chain, match, target))
	}
}

func ListFilterChainsCmd() []string {
	return []string{
		"bash", "-c", `iptables-save \
		| sed -n '/^*filter/,/^[^:]/{/^:/!d;s/:\(\w*\) .*/\1/;p}'`,
	}
}

func (ipcc IPChainCmd) ChainCmd(name string) string {
	return "iptables " + string(ipcc) + " " + name
}

func EnsureChainFunc(runner *Runner, chain string) func() error {
	return func() error {
		if err := runner.Run(ListFilterChainsCmd()); err != nil {
			return err
		}
		if !slices.Contains(strings.Split(runner.Last().Out, "\n"), chain) {
			if err := runner.Line(NEW.ChainCmd(chain)); err != nil {
				return err
			}
		}
		runner.Line(
			CHECK.FilterRule("FORWARD", "--out-interface "+InternetDevice, chain),
		)
		if runner.Last().Code != 0 {
			return runner.Line(
				APPEND.FilterRule("FORWARD", "--out-interface "+InternetDevice, chain),
			)
		}
		return nil
	}
}

func RemoveChainCmdFunc(runner *Runner, chain string) func() error {
	return func() error {
		if err := runner.Run(ListFilterChainsCmd()); err != nil {
			return err
		}
		if slices.Contains(strings.Split(runner.Last().Out, "\n"), chain) {
			err := runner.Line(
				CHECK.FilterRule("FORWARD", "--out-interface "+InternetDevice, chain),
			)
			if err != nil && runner.Last().Code > 1 {
				return err
			}
			if runner.Last().Code == 0 {
				err := runner.Line(
					DELETE.FilterRule("FORWARD", "--out-interface "+InternetDevice, chain),
				)
				if err != nil {
					return err
				}
			}
			if err := runner.Line(
				FLUSH.ChainCmd(chain),
				DELETE_CHAIN.ChainCmd(chain),
			); err != nil {
				return err
			}
		}
		return nil
	}
}
