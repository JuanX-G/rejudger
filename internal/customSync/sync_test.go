package customSync

import (
	"testing"
)

func TestLockedValueRead(t *testing.T) {
	const TEST_INT = 8
	lck := NewLockedValue(TEST_INT)
	var copy int
	lck.WithReadPointer(func(i *int) {
		copy = *i
	})
	if copy != TEST_INT {
		t.Fatalf("wrote: %d as the locked value, found: %d", TEST_INT, copy)
	}
}

func TestLockedValueWrite(t *testing.T) {
	const TEST_INT = 8
	const TEST_INT2 = 7
	lck := NewLockedValue(TEST_INT)
	lck.WithWritePointer(func(i *int) {
		*i = TEST_INT2
	})
	if lck.v != TEST_INT2 {
		t.Fatalf("overwritten the locked value of: %d, with: %d , found: %d", TEST_INT, TEST_INT2, lck.v)
	}
}
