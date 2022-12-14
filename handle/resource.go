package handle

import (
	"github.com/plockc/gateway/resource"
)

type Allowed int

const (
	GET_ALLOWED Allowed = iota
	LIST_ALLOWED
	DELETE_ALLOWED
	UPSERT_ALLOWED
)

type Resources struct {
	Name string
	// the factory will need to parse the ID from a string for URL handling
	Factory       func(ids ...string) (resource.Resource, error)
	Relationships map[string]Resources
	Allowed       []Allowed
}
