package gateway

// what we really want is
// type Pipeline[I any, O any](i I) O
// then have a method on Pipe as a wrapper
// however function types (like above) cannot have generics
func Pipe[I any, O any, Piped any](existing func(I) (Piped, error), next func(Piped) (O, error)) func(I) (O, error) {
	return func(i I) (O, error) {
		p, err := existing(i)
		var o O
		if err != nil {
			return o, err
		}
		return next(p)
	}
}
