// Package config provides a generic config [Option]
package config

import "flag"

type Option[T flag.Value] struct {
	Name        string
	Value       T
	Description string
	Default     string
	Source      string
}

func (receiver Option[T]) Set(s string) error {
	return receiver.Value.Set(s)
}

func (receiver Option[T]) String() string {
	return receiver.Value.String()
}
