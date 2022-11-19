package gateway

import (
	"fmt"
	"strings"
)

type IPSet struct {
	Name    string
	Members []MAC
}

func (ipSet IPSet) String() string {
	return ipSet.Name
}

func (ipSet IPSet) DestroyCmdLine() string {
	return "ipset destroy -exist " + ipSet.Name
}

func (ipSet IPSet) SyncCmdLines() []string {
	name := ipSet.Name
	cmdlines := []string{
		"ipset destroy -exist " + name + "-builder",
		"ipset -N -exist " + name + " hash:mac",
		"ipset -N -exist " + name + "-builder hash:mac",
	}
	for _, m := range ipSet.Members {
		cmdlines = append(cmdlines, "ipset -A "+name+"-builder "+m.String())
	}
	cmdlines = append(
		cmdlines,
		"ipset swap "+name+"-builder "+name,
		"ipset destroy "+name+"-builder",
	)

	return cmdlines
}

func (ipSet IPSet) Match() string {
	return "-m set --match-set " + ipSet.String() + " src"
}

func (ipSet *IPSet) Load(runner *Runner) error {
	err := runner.Line("ipset save -sorted " + ipSet.Name)
	if err != nil {
		return fmt.Errorf("failed to list ipset %s: %w", ipSet.Name, err)
	}
	elems := Keep(strings.Split(runner.LastOut(), "\n"), func(s string) bool {
		return strings.HasPrefix(s, "add ")
	})
	elems = Map(elems, func(s string) string {
		return strings.TrimPrefix(s, "add "+ipSet.Name+" ")
	})
	macs, err := MapErrable(elems, func(s string) (MAC, error) {
		return MACFromString(s)
	})
	if err != nil {
		return fmt.Errorf("failed to parse macs: %w", err)
	}
	ipSet.Members = macs
	return nil
}
