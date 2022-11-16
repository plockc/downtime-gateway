package gateway

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

// AssignFunc creates a function that sets the output
// of the function passed in (after applying input from func 'i') to pointer 'out'
// This is useful for running a Pipeline inside a Do, as everything is deferred execution
func AssignFunc[I any, O any](in func() I, out *O, f func(I) (O, error)) func() error {
	return func() error {
		var err error
		*out, err = f(in())
		return err
	}
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

func LogFn(msg ...any) func() error {
	fmt.Println(msg...)
	return nil
}
