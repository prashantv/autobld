package proxy

import "strings"

// Type represents the a type of proxy.
type Type int

//go:generate stringer -type=Type

// List of different types of proxies.
const (
	// TCP is the default type of proxy, if none is specified.
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

// UnmarshalYAML is used to unmarshal Type from the YAML configuration.
func (t *Type) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*t = stringToType(s)
	return nil
}
