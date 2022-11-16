package gateway

type NS string

func (ns NS) WrapCmd(cmd []string) []string {
	return append(
		[]string{"ip", "netns", "exec", string(ns)}, cmd...,
	)
}

func (ns NS) WrapCmdLine(cmd string) string {
	return "ip netns exec " + string(ns) + " " + cmd
}

func (ns NS) WrapCmdLines(cmds []string) []string {
	return Multiline(cmds).Map(ns.WrapCmdLine)
}

func (ns NS) DelCmd() []string {
	delCmd := "ip netns del " + string(ns)
	nsExists := "ip netns list | grep -q " + string(ns)
	return []string{"sh", "-c", `if ` + nsExists + `; then ` + delCmd + `; fi`}
}

func (ns NS) CreateCmd() string {
	return "ip netns add " + string(ns)
}
