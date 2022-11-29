package resource

import "fmt"

type FailUnimplementedMethods struct {
}

func (FailUnimplementedMethods) Delete() error {
	return fmt.Errorf("delete unimplemented")
}

func (FailUnimplementedMethods) Create() error {
	return fmt.Errorf("create unimplemented")
}

func (FailUnimplementedMethods) List() ([]string, error) {
	return nil, fmt.Errorf("list unimplemented")
}

func (FailUnimplementedMethods) Clear() error {
	return fmt.Errorf("clear unimplemented")
}
