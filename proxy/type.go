package proxy

import "strings"

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

// TODO: use stringer
func (t Type) String() string {
	switch t {
	case TCP:
		return "TCP"
	case HTTP:
		return "HTTP"
	default:
		return "Unknown"
	}
}

func (t *Type) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*t = stringToType(s)
	return nil
}
