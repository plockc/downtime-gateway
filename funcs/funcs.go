package funcs

import "fmt"

// Keep returns a new slice of items that return true for filter
func Keep[S ~[]T, T any](items S, keepFunc func(t T) bool) []T {
	kept := []T{}
	for _, i := range items {
		if keepFunc(i) {
			kept = append(kept, i)
		}
	}
	return kept
}

// Map returns a new slice of items
func Map[I, O any](items []I, mapFunc func(i I) O) []O {
	mapped := make([]O, len(items))
	for i, e := range items {
		mapped[i] = mapFunc(e)
	}
	return mapped
}

// MapErrable returns a new slice of items
func MapErrable[I, O any](items []I, mapFunc func(i I) (O, error)) ([]O, error) {
	mapped := make([]O, len(items))
	for i, e := range items {
		var err error
		mapped[i], err = mapFunc(e)
		if err != nil {
			return nil, err
		}
	}
	return mapped, nil
}

func Do(fs ...func() error) error {
	for _, f := range fs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func ExpectFailFunc(msg string, f func() error) func() error {
	return func() error {
		if err := f(); err == nil {
			return fmt.Errorf("expected failure for: %s", msg)
		}
		return nil
	}
}

// AssignFunc will take a func returning a value and error such that it
// only returns an error and the value is assigned to a pointer
func AssignFunc[T any](f func() (T, error), target *T) func() error {
	return func() error {
		var err error
		*target, err = f()
		return err
	}
}
