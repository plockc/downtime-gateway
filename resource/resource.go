package resource

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Resource interface {
	Id() string
	// Create returns true if created vs already existing
	Create() error
	// Delete returns true if delete vs already gone
	Delete() error
	List() ([]string, error)
	Clear() error
}

type Loader interface {
	Load() error
}

type Lifecycle struct {
	Resource
}

func NewLifecycle(r Resource) Lifecycle {
	return Lifecycle{Resource: r}
}

func (lf Lifecycle) Exists() (bool, error) {
	list, err := lf.List()
	if err != nil {
		return false, err
	}
	return slices.Contains(list, lf.Resource.Id()), nil
}

func (lf Lifecycle) Ensure() (bool, error) {
	if exists, err := lf.Exists(); err != nil {
		return false, fmt.Errorf("cannot ensure %v exists: %w", lf.Resource.Id(), err)
	} else if exists {
		return false, nil
	}
	return true, lf.Create()
}

func (lf Lifecycle) EnsureDeleted() (bool, error) {
	if exists, err := lf.Exists(); err != nil {
		return false, fmt.Errorf("cannot ensure %v is deleted: %w", lf.Resource.Id(), err)
	} else if !exists {
		return false, nil
	}
	return true, lf.Delete()
}

func (lf Lifecycle) EnsureCleared() (bool, error) {
	list, err := lf.List()
	if err != nil {
		return false, fmt.Errorf("cannot ensure %T is deleted: %w", lf.Resource, err)
	}
	if len(list) == 0 {
		return false, nil
	}
	return true, lf.Clear()
}
