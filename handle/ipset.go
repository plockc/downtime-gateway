package handle

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/plockc/gateway"
	"golang.org/x/exp/slices"
)

var SetsHandlers = MethodHandlers{
	http.MethodGet: TypedHandler{
		Handler: setGet,
		T:       reflect.TypeOf(gateway.IPSet{}),
	},
	http.MethodPut: TypedHandler{
		Handler: setPut,
		T:       reflect.TypeOf(gateway.IPSet{}),
	},
	//http.MethodPatch:  setPatch,
	http.MethodDelete: TypedHandler{
		Handler: setDelete,
		T:       reflect.TypeOf(gateway.IPSet{}),
	},
}

func setDelete(parts []string, _ any) (int, any, error) {
	if len(parts) == 0 || parts[0] == "" {
		return 400, nil, fmt.Errorf("missing set name")
	}
	if len(parts) > 3 {
		return 400, nil, fmt.Errorf("can only delete sets or member of sets")
	}
	if len(parts) == 1 {
		err := Runner.Line("ipset destroy -exist " + parts[0])
		if err != nil {
			return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
		}
		return 200, nil, nil
	}
	if parts[1] != "members" {
		return 400, nil, fmt.Errorf("can only delete 'members' from sets")
	}

	ipSet := gateway.IPSet{Name: parts[0]}
	err := ipSet.Load(Runner)
	if err != nil {
		return 500, nil, err
	}
	if len(ipSet.Members) == 0 {
		return 200, nil, nil
	}

	if len(parts) == 2 || parts[2] == "" {
		err := Runner.Line("ipset flush " + parts[0])
		if err != nil {
			return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
		}
		return 204, nil, nil
	}

	mac, err := gateway.MACFromString(parts[2])
	if err != nil {
		return 400, nil, fmt.Errorf("could not convert '" + parts[2] + " into a MAC")
	}

	if !slices.Contains(ipSet.Members, mac) {
		return 200, nil, nil
	}

	err = Runner.Line("ipset del " + parts[0] + " " + parts[2])
	if err != nil {
		return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
	}
	return 204, nil, nil
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

	setName := parts[0]
	setExisted, err := ipSetExists(setName)
	if err != nil {
		return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
	}

	successfulReturnCode := 200
	if !setExisted {
		successfulReturnCode = 201
	}

	if len(parts) == 1 {
		// PUT the IP Set then return
		err := Runner.Line("ipset -N -exist " + setName + " hash:mac")
		if err != nil {
			return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
		}
		return successfulReturnCode, nil, nil
	}

	// will be PUTing a member
	ipSet := gateway.IPSet{Name: setName}
	if err := ipSet.Load(Runner); err != nil {
		return 500, nil, err
	}

	mac, err := gateway.MACFromString(parts[2])
	if err != nil {
		return 500, nil, err
	}

	if slices.Contains(ipSet.Members, mac) {
		return 200, nil, nil
	}

	err = Runner.Line("ipset add " + setName + " " + parts[2])
	if err != nil {
		return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
	}
	return 201, nil, nil
}

func ipSets() ([]string, error) {
	err := Runner.Line("ipset list -n")
	if err != nil {
		return nil, fmt.Errorf("failed to list ipsets: %w", err)
	}
	return strings.Split(Runner.LastOut(), "\n"), nil
}

func ipSetExists(name string) (bool, error) {
	sets, err := ipSets()
	if err != nil {
		return false, err
	}
	return slices.Contains(sets, name), nil
}

func setGet(parts []string, _ any) (int, any, error) {
	if len(parts) > 3 {
		return 400, nil, fmt.Errorf("too many path elements in URL")
	}

	sets, err := ipSets()
	if err != nil {
		return 500, nil, fmt.Errorf("failed to list ipsets: %w", err)
	}

	// could be simple list the sets "/sets"
	if len(parts) == 0 {
		return 200, sets, nil
	}

	setName := parts[0]

	if len(parts) >= 2 && parts[1] != "members" {
		return 400, nil, fmt.Errorf("can only get 'members' from sets")
	}

	// handle getting a single member (an exists test, either 200 or 404)
	// /sets/foo/members/12:12:12:12:12:12
	if len(parts) == 3 {
		err := Runner.Line("ipset test " + setName + " " + parts[2])
		if err == nil {
			return 200, nil, nil
		}
		if strings.Contains(Runner.Last().Out, "NOT in set") {
			return 404, nil, fmt.Errorf("missing " + parts[2])
		}
		return 500, nil, fmt.Errorf("%w %s", err, Runner.LastOut())
	}

	// handle Listing the members of a given set
	// /sets/foo/members
	if !slices.Contains(sets, setName) {
		return 404, nil, fmt.Errorf("missing set " + setName)
	}

	ipSet := gateway.IPSet{Name: setName}
	err = ipSet.Load(Runner)
	if err != nil {
		return 500, nil, err
	}

	return 200, ipSet.Members, nil
}
