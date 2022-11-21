package iptables

import (
	"strings"

	"github.com/plockc/gateway/runner"
	"golang.org/x/exp/slices"
)

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

func ListFilterChainsCmd() []string {
	return []string{
		"bash", "-c", `iptables-save \
		| sed -n '/^*filter/,/^[^:]/{/^:/!d;s/:\(\w*\) .*/\1/;p}'`,
	}
}

func EnsureChainFunc(runner *runner.Runner, chain string) func() error {
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

func RemoveChainCmdFunc(runner *runner.Runner, chain string) func() error {
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
