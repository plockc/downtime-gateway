package handle

import (
	"fmt"

	"github.com/plockc/gateway/resource"
)

var Versions = Resources{
	Name: "Version",
	Factory: func(ids ...string) (resource.Resource, error) {
		switch len(ids) {
		case 0:
			return Version{}, nil
		default:
			if ids[0] != "v1" {
				return nil, fmt.Errorf("only v1 supported")
			}
			return Version{Name: ids[0]}, nil
		}
	},
	Relationships: map[string]Resources{
		"netns": Namespaces,
	},
	Allowed: []Allowed{GET_ALLOWED, LIST_ALLOWED},
}

type Version struct {
	resource.FailUnimplementedMethods
	Name string
}

var _ resource.Resource = Version{}

func (v Version) Id() string {
	return v.Name
}

func (v Version) List() ([]string, error) {
	return []string{"v1"}, nil
}
