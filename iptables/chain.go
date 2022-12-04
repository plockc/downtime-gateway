package iptables

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/resource"
)

var chainRegex = regexp.MustCompile(`\w+ (\w+) .*`)

type Chain struct {
	Name  string `json:"-"`
	Table `json:"-"`
}

func NewChain(table Table, name string) Chain {
	return Chain{Name: name, Table: table}
}

func (c Chain) ChainResource() *ChainRes {
	return &ChainRes{Chain: c}
}

func (c Chain) String() string {
	return c.Table.String() + ":chain[" + c.Name + "]"
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
	return chain.Runner().RunLine(DELETE_CHAIN.ChainCmd(chain.Id()))
}

func (chain ChainRes) Create() error {
	return chain.Runner().RunLine(NEW.ChainCmd(chain.Id()))
}

func (chain ChainRes) List() ([]string, error) {
	run := chain.Runner()
	results, err := run.ExecLine("iptables -t " + chain.Table.Name + " -L")
	if err != nil {
		return nil, err
	}
	chainLines := funcs.Keep(strings.Split(results.Out, "\n"), func(s string) bool {
		return strings.HasPrefix(s, "Chain")
	})
	return funcs.Map(chainLines, func(s string) string {
		chainMatches := chainRegex.FindStringSubmatch(s)
		if len(chainMatches) != 2 {
			err = fmt.Errorf("did not find chain name in '%s'", s)
			return ""
		}
		return chainMatches[1]
	}), err
}

func (chain ChainRes) Clear() error {
	return chain.Runner().RunLine("iptables -X")
}
