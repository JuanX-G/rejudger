package testingservices

import "os"

type MockLogDest struct {
	logged      [][]byte
	loggedCount int
	Closed      bool
}

func (ld *MockLogDest) Write(p []byte) (n int, err error) {
	if ld.Closed == true {
		return 0, os.ErrClosed
	}
	ld.logged = append(ld.logged, p)
	ld.loggedCount++
	return len(p), nil
}

func (ld *MockLogDest) Close() error {
	ld.Closed = true
	return nil
}

