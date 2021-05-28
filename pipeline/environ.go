package pipeline

import (
	"os"
	"strings"
)

type Environ struct {
	items map[string]string
}

func NewEnviron() *Environ {
	res := Environ{
		items: map[string]string{},
	}

	for _, l := range os.Environ() {
		kv := strings.SplitN(l, "=", 2)
		if len(kv) < 2 {
			continue
		}

		res.items[kv[0]] = kv[1]
	}

	return &res
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
func (e Environ) Getenv(key string) string {
	return e.items[key]
}

// Setenv sets the value of the environment variable named by the key.
// It returns an error, if any.
func (e *Environ) Setenv(key, value string) {
	e.items[key] = value
}

// Unsetenv unsets a single environment variable.
func (e *Environ) Unsetenv(key string) {
	delete(e.items, key)
	os.Environ()
}

// Environ returns a copy of strings representing the environment,
// in the form "key=value".
func (e *Environ) Environ() []string {
	res := make([]string, len(e.items))

	i := 0
	for k, v := range e.items {
		res[i] = k + "=" + v
		i++
	}
	return res
}

// Expand replaces ${var} or $var in the string according to the values
// of the environment variables. References to undefined
// variables are replaced by the empty string.
func (e Environ) Expand(str string) string {
	return os.Expand(str, func(key string) string { return e.items[key] })
}
