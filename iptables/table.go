package iptables

import (
	"fmt"

	"github.com/plockc/gateway/resource"
)

var _ resource.Resource = TableRes{}

type Table struct {
	Name string
	resource.NS
}

type TableRes struct {
	Table
	resource.FailUnimplementedMethods
}

func NewTable(ns resource.NS, name string) Table {
	return Table{Name: name, NS: ns}
}

func (t TableRes) Id() string {
	return t.Name
}

func (t TableRes) List() ([]string, error) {
	return []string{"filter"}, nil
}

func (TableRes) Clear() error {
	return fmt.Errorf("unimplemented")
}
