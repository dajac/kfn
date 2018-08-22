package flag

import (
	"errors"
	"fmt"
	"strings"
)

type Config map[string]string

func (i *Config) String() string {
	return fmt.Sprint(*i)
}

func (i *Config) Set(value string) error {
	res := strings.Split(value, ":")

	if len(res) != 2 {
		return errors.New("keyvalue flag must be key:value")
	}

	(*i)[res[0]] = res[1]

	return nil
}
