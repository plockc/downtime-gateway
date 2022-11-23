package resource

import "fmt"

type FailUnimplementedMethods struct {
}

func (FailUnimplementedMethods) Delete() error {
	return fmt.Errorf("unimplemented")
}

func (FailUnimplementedMethods) Create() error {
	return fmt.Errorf("unimplemented")
}

func (FailUnimplementedMethods) List() ([]string, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (FailUnimplementedMethods) Clear() error {
	return fmt.Errorf("unimplemented")
}
