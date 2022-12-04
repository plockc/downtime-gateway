package iptables

import (
	"fmt"

	"github.com/plockc/gateway/resource"
)

var _ resource.Resource = TableRes{}

type Table struct {
	Name        string `json:"-"`
	resource.NS `json:"-"`
}

func FilterTable(ns resource.NS) Table {
	return Table{Name: "filter", NS: ns}
}

func NewTable(ns resource.NS, name string) Table {
	return Table{Name: name, NS: ns}
}

func (t Table) TableResource() TableRes {
	return NewTableResource(t)
}

type TableRes struct {
	Table
	resource.FailUnimplementedMethods
}

func NewTableResource(t Table) TableRes {
	return TableRes{Table: t}
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
