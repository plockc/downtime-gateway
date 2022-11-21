package funcs

import "fmt"

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

func LogFn(msg ...any) func() error {
	fmt.Println(msg...)
	return nil
}

// PipeFunc creates a function that sets the output
// of the function passed in (after applying input from func 'i') to pointer 'out'
// This is useful for running a Pipeline inside a Do, as everything is deferred execution
func PipeFunc[I any, O any](in func() I, out *O, f func(I) (O, error)) func() error {
	return func() error {
		var err error
		*out, err = f(in())
		return err
	}
}
