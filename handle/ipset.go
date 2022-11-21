package handle

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/plockc/gateway/address"
	"github.com/plockc/gateway/iptables"
	"github.com/plockc/gateway/resource"
	"golang.org/x/exp/slices"
)

var SetsHandlers = MethodHandlers{
	http.MethodGet: TypedHandler{
		Handler: setGet,
		T:       reflect.TypeOf(iptables.IPSet{}),
	},
	http.MethodPut: TypedHandler{
		Handler: setPut,
		T:       reflect.TypeOf(iptables.IPSet{}),
	},
	//http.MethodPatch:  setPatch,
	http.MethodDelete: TypedHandler{
		Handler: setDelete,
		T:       reflect.TypeOf(iptables.IPSet{}),
	},
}

func setDelete(parts []string, _ any) (int, any, error) {
	if len(parts) == 0 || parts[0] == "" {
		return 400, nil, fmt.Errorf("missing set name")
	}
	if len(parts) > 3 {
		return 400, nil, fmt.Errorf("can only delete sets or member of sets")
	}
	ipSet := iptables.NewIPSet(parts[0], NS)
	ipSetLC := resource.Lifecycle[string]{Resource: ipSet}
	if len(parts) == 1 {
		deleted, err := ipSetLC.EnsureDeleted()
		if err != nil {
			return 500, nil, fmt.Errorf("%w %s", err, ipSet.Runner.LastOut())
		}
		if deleted {
			return 204, nil, nil
		}
		return 200, nil, nil
	}
	if parts[1] != "members" {
		return 400, nil, fmt.Errorf("can only delete 'members' from sets")
	}

	macs, err := (&iptables.Member{IPSet: ipSet}).List()
	if err != nil {
		return 500, nil, err
	}
	if len(macs) == 0 {
		return 200, nil, nil
	}

	if len(parts) == 2 || parts[2] == "" {
		err := ipSet.Clear()
		if err != nil {
			return 500, nil, fmt.Errorf("%w %s", err, ipSet.Runner.LastOut())
		}
		return 204, nil, nil
	}

	mac, err := address.MACFromString(parts[2])
	if err != nil {
		return 400, nil, fmt.Errorf("could not convert '" + parts[2] + " into a MAC")
	}

	if !slices.Contains(macs, mac) {
		return 200, nil, nil
	}

	member := iptables.NewMember(ipSet, mac)
	memberLifecycle := resource.Lifecycle[address.MAC]{
		Resource: member,
	}

	deleted, err := memberLifecycle.EnsureDeleted()
	if err != nil {
		return 500, nil, fmt.Errorf("%w %s", err, member.IPSet.Runner.LastOut())
	}
	if deleted {
		return 204, nil, nil
	}
	return 200, nil, nil
}

func setPut(parts []string, data any) (int, any, error) {
	switch {
	case data != nil:
		return 400, nil, fmt.Errorf("not expecting any body")
	case len(parts) == 0 || len(parts[0]) == 0:
		return 400, nil, fmt.Errorf("cannot PUT without a set name")
	case len(parts) == 1 && parts[0] == "":
		return 400, nil, fmt.Errorf("cannot PUT without a set name")
	case len(parts) == 3 && parts[1] != "members":
		return 400, nil, fmt.Errorf("cannot PUT to a set for anything other than members")
	case len(parts) == 2:
		return 400, nil, fmt.Errorf("cannot PUT to a set without members/MAC")
	case len(parts) > 3:
		return 400, nil, fmt.Errorf("URL path too long")
	}

	ipSet := iptables.NewIPSet(parts[0], NS)
	ipSetLifecycle := resource.Lifecycle[string]{Resource: ipSet}
	if len(parts) == 1 {
		// Only creating the IPSet
		created, err := ipSetLifecycle.Ensure()
		if err != nil {
			return 500, nil, fmt.Errorf("%w %s", err, ipSet.Runner.LastOut())
		}
		if created {
			return 201, nil, nil
		}
		return 200, nil, nil
	}

	// will be PUTing a member
	mac, err := address.MACFromString(parts[2])
	if err != nil {
		return 500, nil, err
	}

	member := iptables.NewMember(ipSet, mac)
	memberLC := resource.Lifecycle[address.MAC]{Resource: member}
	created, err := memberLC.Ensure()
	if err != nil {
		return 500, nil, fmt.Errorf("%w %s", err, ipSet.Runner.LastOut())
	}
	if created {
		return 201, nil, nil
	}
	return 200, nil, nil
}

func setGet(parts []string, _ any) (int, any, error) {
	if len(parts) > 3 {
		return 400, nil, fmt.Errorf("too many path elements in URL")
	}

	sets, err := iptables.NewIPSet(parts[0], NS).List()
	if err != nil {
		return 500, nil, fmt.Errorf("failed to list ipsets: %w", err)
	}

	// could be simple list the sets "/sets"
	if len(parts) == 0 {
		return 200, sets, nil
	}

	setName := parts[0]
	ipSet := iptables.NewIPSet(setName, NS)

	if len(parts) >= 2 && parts[1] != "members" {
		return 400, nil, fmt.Errorf("can only get 'members' from sets")
	}

	// handle getting a single member (an exists test, either 200 or 404)
	// /sets/foo/members/12:12:12:12:12:12
	if len(parts) == 3 {
		mac, err := address.MACFromString(parts[2])
		if err != nil {
			return 500, nil, err
		}
		member := iptables.NewMember(ipSet, mac)
		memberLC := resource.Lifecycle[address.MAC]{Resource: member}
		exists, err := memberLC.Exists()
		if err != nil {
			return 500, nil, err
		}
		if !exists {
			return 404, nil, fmt.Errorf("missing " + parts[2])
		}
		return 200, nil, nil
	}

	// handle Listing the members of a given set
	// /sets/foo/members
	if !slices.Contains(sets, setName) {
		return 404, nil, fmt.Errorf("missing set " + setName)
	}

	members, err := iptables.NewMember(ipSet, address.MAC{}).List()
	if err != nil {
		return 500, nil, err
	}

	return 200, members, nil
}
