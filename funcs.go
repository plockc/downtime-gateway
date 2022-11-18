package gateway

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
