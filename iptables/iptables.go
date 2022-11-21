package iptables

import "github.com/plockc/gateway/runner"

type IPRuleCmd string

const (
	DROP   = "DROP"
	RETURN = "RETURN"

	APPEND IPRuleCmd = "-A"
	CHECK  IPRuleCmd = "-C"
	DELETE IPRuleCmd = "-D"
)

func (iptc IPRuleCmd) FilterRule(chain, match, target string) string {
	return "iptables " + string(iptc) + " " + chain + " " + match + " -j " + target
}

func EnsureIPRuleFunc(runner runner.Runner, chain, match, target string) func() error {
	return func() error {
		err := runner.Line(CHECK.FilterRule(chain, match, target))
		if err != nil {
			return err
		}
		return runner.Line(APPEND.FilterRule(chain, match, target))
	}
}
