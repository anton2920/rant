package main

import "unsafe"

type SyncQueue struct {
	M     Mutex
	Items []unsafe.Pointer
	Pos   int
}

func (sq *SyncQueue) Get() unsafe.Pointer {
	var ret unsafe.Pointer

	sq.M.Lock()
	if len(sq.Items) > sq.Pos {
		ret = sq.Items[sq.Pos]
		sq.Pos++
	} else {
		sq.Pos = 0
		sq.Items = sq.Items[:0]
	}
	sq.M.Unlock()

	return ret
}

func (sq *SyncQueue) Put(item unsafe.Pointer) {
	sq.M.Lock()
	sq.Items = append(sq.Items, item)
	sq.M.Unlock()
}
