package main

/* Mutex implements spinwait lock. */
type Mutex struct {
	State int32
}

const (
	MutexUnlocked = iota
	MutexLocked   = 1
)

const enabled = false

//go:noescape
//go:nosplit
func Cas32(val *int32, old, new int32) bool

//go:nosplit
func Pause()

func (m *Mutex) Lock() {
	if !enabled {
		return
	}
	for !Cas32(&m.State, MutexUnlocked, MutexLocked) {
		Pause()
	}
}

func (m *Mutex) Unlock() {
	if !enabled {
		return
	}
	if m.State == MutexUnlocked {
		panic("Lock() on locked mutex")
	}

	for !Cas32(&m.State, MutexLocked, MutexUnlocked) {
		Pause()
	}
}
