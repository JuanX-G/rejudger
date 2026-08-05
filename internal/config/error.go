package config

import (
	"fmt"
)

//go:generate stringer -type=ConfigErrorType -trimprefix=Config
type ConfigErrorType int
const (
	ConfigErrorOpenFile ConfigErrorType = iota
	ConfigErrorAtoiError
	ConfigErrorDecodeError
)

type ConfigError struct {
	errType ConfigErrorType
	msg string
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("config parsing error. Type: %s, message: %s", e.errType.String(), e.msg)
}
