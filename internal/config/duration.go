package config

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

//go:generate stringer -type=DeadlineParsingErrorType -trimprefix=ConfigParsing
type DeadlineParsingErrorType int
const (
	ConfigParsingErrorUnexpectedLengthSpecifier DeadlineParsingErrorType = iota
	ConfigParsingErrorAtoiError
)

type DeadlineParsingError struct {
	errType DeadlineParsingErrorType
	msg string
}

func (e DeadlineParsingError) Error() string {
	return fmt.Sprintf("config parsing error. Type: %s, message: %s", e.errType.String(), e.msg)
}

type StageDeadline struct {
	FrequencyStr string `yaml:"frequency"`
	frequencyDays int
	frequencHours int
	frequencyDayOfTheWeek int
	frequencyDayOfTheMonth int
}

func(s *StageDeadline) ParseDeadline() error {
	if len(s.FrequencyStr) >= 2 {
		if s.FrequencyStr[0] == 'D' {
			daysStr := strings.Builder{}
			hourStr := strings.Builder{}
			for i, c := range s.FrequencyStr {
				if !unicode.IsNumber(c) {
					if c == 'H' {
						for j := i; j < len(s.FrequencyStr); j++ {
							hourStr.WriteByte(s.FrequencyStr[j])
						}
						break
					} else {
						return DeadlineParsingError{msg: fmt.Sprintf("specifier: %c found, but only 'H' is allowed here", c), errType: ConfigParsingErrorUnexpectedLengthSpecifier}
					}
				} else {
					daysStr.WriteRune(c)
				}
			}
			days, err := strconv.Atoi(daysStr.String())
			if err != nil {
				return DeadlineParsingError{msg: fmt.Sprintf("string parsing error occured, found: %s. Atoi reported: %s", daysStr.String(), err.Error()), errType: ConfigParsingErrorAtoiError}
			}
			s.frequencyDays = days
			var hours int
			if hourStr.Len() > 0 {
				hours, err = strconv.Atoi(hourStr.String())
				if err != nil {
					return DeadlineParsingError{msg: fmt.Sprintf("string parsing error occured, found: %s. Atoi reported: %s", hourStr.String(), err.Error()), errType: ConfigParsingErrorAtoiError}
				}
			}
			s.frequencHours = hours
		} else if s.FrequencyStr[0] == 'W' {
			day, err := strconv.Atoi(string(s.FrequencyStr[1]))
			if err != nil {
				return err
			}
			s.frequencyDayOfTheWeek = day
		}
	}
	return nil
}
