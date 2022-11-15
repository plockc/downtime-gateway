package gateway

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

func (ipSet IPSet) CreateCmdLines() []string {
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

func (ipSet IPSet) BlockInternetCmdLine() string {
	return ("iptables -A FORWARD" +
		" -m set --match-set " + ipSet.String() + " src" +
		" -j DROP")
}

func (ipSet IPSet) ReturnInternetCmdLine() string {
	return ("iptables -A FORWARD" +
		" -m set --match-set " + ipSet.String() + " src" +
		" -j RETURN")
}
