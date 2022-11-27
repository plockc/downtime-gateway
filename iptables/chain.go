package iptables

import (
	"strings"

	"github.com/plockc/gateway/resource"
	"golang.org/x/exp/slices"
)

type Chain struct {
	Name  string
	Table `json:"-"`
}

func NewChain(table Table, name string) Chain {
	return Chain{Name: name, Table: table}
}

var _ resource.Resource = ChainRes{}

type ChainRes struct {
	resource.FailUnimplementedMethods
	Chain
}

func (chain ChainRes) Id() string {
	return chain.Name
}

func (chain ChainRes) Delete() error {
	return chain.Runner().Line(DELETE_CHAIN.ChainCmd(chain.Id()))
}

func (chain ChainRes) Create() error {
	return chain.Runner().Line(NEW.ChainCmd(chain.Id()))
}

func ListFilterChainsCmd() []string {

	return []string{
		"bash", "-c", `iptables-save \
		| sed -n '/^*filter/,/^[^:]/{/^:/!d;s/:\(\w*\) .*/\1/;p}'`,
	}
}

func EnsureChainFunc(runner *resource.Runner, chain string) func() error {
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

func RemoveChainCmdFunc(runner *resource.Runner, chain string) func() error {
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
