package handle

import "github.com/plockc/gateway/resource"

type Allowed int

const (
	GET_ALLOWED Allowed = iota
	LIST_ALLOWED
	DELETE_ALLOWED
	UPSERT_ALLOWED
)

type Factory func(string) (resource.Resource, error)
type ChainedFactory func() (ChainedFactory, Factory)

type Resources struct {
	Label string
	// the factory will need to parse the ID from a string for URL handling
	ChainedFactory
	Relationships map[string]Resources
	Allowed       []Allowed
}
