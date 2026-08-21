package sharedtesting

import (
	"testing"
)

func ServiceSetupFail(tb testing.TB, err error, service string) {
	tb.Fatalf("error: %s occured at service: %s, startup, could not continue testing", err, service)
}

func OperationFail(tb testing.TB, err error, operation string) {
	tb.Fatalf("error: %s occured performing: %s, could not continue testing", err, operation)
}

func WrongHttpCode(tb testing.TB, expected int, found int) {
	tb.Fatalf("error: expected http status code: %d, found: %d", expected, found)
}
