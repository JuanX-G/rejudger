package logger

import "strings"

type LogLevel int

//go:generate stringer -type=LogLevel -trimprefix=LogLevel
const (
	LogLevelInfo LogLevel = iota
	LogLevelDebug
	LogLevelError
)

type LogMessage struct {
	Level  LogLevel
	Source string
	Msg    string
}

func (lm *LogMessage) String() string {
	sb := strings.Builder{}
	sb.Grow(len(lm.Msg) + len(lm.Source) + 8)
	sb.WriteString("level = ")
	sb.WriteString(lm.Level.String())

	sb.WriteString("| source = ")
	sb.WriteString(lm.Source)

	sb.WriteString("| message = ")
	sb.WriteString(lm.Msg)
	return sb.String()
}

func (lm *LogMessage) Bytes() []byte {
	return []byte(lm.String())
}
