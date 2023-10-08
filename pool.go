package main

import "unsafe"

type SyncPool struct {
	M     Mutex
	Items []unsafe.Pointer

	New func() unsafe.Pointer
}

func NewSyncPool(size int, newF func() unsafe.Pointer) *SyncPool {
	ret := new(SyncPool)
	ret.Items = make([]unsafe.Pointer, 0, size)
	ret.New = newF
	return ret
}

func (sp *SyncPool) Get() unsafe.Pointer {
	var item unsafe.Pointer

	sp.M.Lock()
	if len(sp.Items) > 0 {
		item = sp.Items[len(sp.Items)-1]
		sp.Items = sp.Items[:len(sp.Items)-1]
		sp.M.Unlock()
	} else {
		sp.M.Unlock()
		item = sp.New()
	}
	return item
}

func (sp *SyncPool) Put(item unsafe.Pointer) {
	sp.M.Lock()
	sp.Items = append(sp.Items, item)
	sp.M.Unlock()
}
