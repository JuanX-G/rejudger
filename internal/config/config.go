package config

import (
	"strconv"
	"unicode"
	"strings"
)

type Pipeline struct {
	Name string `yaml:"name"`
	Stages []PipelineStage `yaml:"stages"`
}

type PipelineStage struct {
	Name string `yaml:"name"`
	PlusOnesRequired int `yaml:"plus_ones_required"`
	Blind bool `yaml:"blind"`
	SoftVeto bool `yaml:"soft_veto"`
	HasDeadline bool `yaml:"has_deadline"`
	Deadline StageDeadline `yaml:"deadline"`
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
						for j := i; i < len(s.FrequencyStr); j++ {
							hourStr.WriteByte(s.FrequencyStr[j])
						}
						break
					} else {
						break // TODO: error
					}
				} else {
					daysStr.WriteRune(c)
				}
			}
			days, err := strconv.Atoi(daysStr.String())
			if err != nil {
				return err
			}
			s.frequencyDays = days
			var hours int
			if hourStr.Len() > 0 {
				hours, err = strconv.Atoi(hourStr.String())
				if err != nil {
					return err
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
