package handle

import (
	"fmt"

	"github.com/plockc/gateway/resource"
)

var Versions = Resources{
	Name: "Version",
	Factory: func(bodyIgnored []byte, ids ...string) (resource.Resource, error) {
		switch len(ids) {
		case 0:
			return Version{}, nil
		default:
			if ids[0] != "v1" {
				return nil, fmt.Errorf("only v1 supported")
			}
			return Version{Named: resource.Named(ids[0])}, nil
		}
	},
	T: nil,
	Relationships: map[string]Resources{
		"netns": Namespaces,
	},
	Allowed: []Allowed{GET_ALLOWED, LIST_ALLOWED},
}

type Version struct {
	resource.FailUnimplementedMethods
	resource.Named
}

var _ resource.Resource = Version{}
