package logger

import (
	"fmt"
	"io"

	"github.com/JuanX-G/crossbow"
)

type LogHandler struct {
	Dest io.WriteCloser
}

func (lh *LogHandler) Init() error {
	if lh.Dest == nil {
		return fmt.Errorf("destination cannot be nil")
	}
	return nil
}

func (lh *LogHandler) Handle(msg crossbow.ContextMessage[LogMessage, any]) (any, error) {
	bytes := msg.Value.Bytes()
	_, err := lh.Dest.Write(bytes)
	return nil, err
}

func (lh *LogHandler) Terminate(err error) {
	lh.Dest.Close()
}
