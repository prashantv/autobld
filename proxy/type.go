package proxy

import "strings"

//go:generate stringer -type=Type
type Type int

const (
	// Default type if not specified is TCP.
	TCP Type = iota
	HTTP
	UnknownType
)

func stringToType(p string) Type {
	p = strings.ToLower(p)
	switch p {
	case "", "tcp":
		return TCP
	case "http":
		return HTTP
	}
	return UnknownType
}

func (t *Type) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*t = stringToType(s)
	return nil
}
