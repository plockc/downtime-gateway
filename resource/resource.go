package resource

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Lifecycle[ID_T comparable] struct {
	Resource[ID_T]
}

type Resource[ID_T comparable] interface {
	fmt.Stringer
	Id() ID_T
	// Create returns true if created vs already existing
	Create() error
	// Delete returns true if delete vs already gone
	Delete() error
	List() ([]ID_T, error)
}

func (lf Lifecycle[ID_T]) Exists() (bool, error) {
	list, err := lf.List()
	if err != nil {
		return false, err
	}
	return slices.Contains(list, lf.Resource.Id()), nil
}

func (lf Lifecycle[ID_T]) Ensure() (bool, error) {
	if exists, err := lf.Exists(); err != nil {
		return false, fmt.Errorf("cannot ensure %v exists: %w", lf.Resource.Id(), err)
	} else if exists {
		return false, nil
	}
	return true, lf.Create()
}

func (lf Lifecycle[ID_T]) EnsureDeleted() (bool, error) {
	if exists, err := lf.Exists(); err != nil {
		return false, fmt.Errorf("cannot ensure %v is deleted: %w", lf.Resource.Id(), err)
	} else if !exists {
		return false, nil
	}
	return true, lf.Delete()
}
