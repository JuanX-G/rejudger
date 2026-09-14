package customSync

import "testing"

func TestMapInit(t *testing.T) {
	mp := NewSafeMap[int, struct{}]()
	if mp.base == nil {
		t.Fatal("SafeMap backing map found to be nil")
	}
}

func TestMapNonExistentKey(t *testing.T) {
	mp := NewSafeMap[int, struct{}]()
	const KEY = 2

	var ran bool
	var found bool
	check := func(fnName string) {
		if found {
			t.Fatalf("%s returned: %t, but key: %d was not stored in the map", fnName, found, KEY)
		} else if ran {
			t.Fatal("WithLock ran the function, but the item was not found")
		}
	}

	found = mp.WithLock(KEY, func(struct{}) {
		ran = true
	})
	check("WithLock")

	found = mp.WithRLock(KEY, func(struct{}) {
		ran = true
	})
	check("WithRLock")
}

func TestMapRetrival(t *testing.T) {
	mp := NewSafeMap[int, int]()
	const KEY = 2
	const VAL = 3
	mp.Store(KEY, VAL)

	var ran bool
	var found bool
	retrived := 0
	check := func(fnName string) {
		if !found {
			t.Fatalf("%s returned: %t, but key: %d was stored in the map", fnName, found, KEY)
		} else if !ran {
			t.Fatalf("%s didn't run the function, but the item was found", fnName)
		} else if retrived != VAL {
			t.Fatalf("%s returned: %d, when: %d was expected", fnName, retrived, VAL)
		}
	}

	found = mp.WithLock(KEY, func(v int) {
		ran = true
		retrived = v
	})
	check("WithLock")

	found = mp.WithRLock(KEY, func(v int) {
		ran = true
		retrived = v
	})
	check("WithRLock")
}
