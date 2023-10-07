package main

type SyncPool[T any] struct {
	M     Mutex
	Items []*T

	New func() *T
}

func NewSyncPool[T any](size int, newF func() *T) *SyncPool[T] {
	ret := new(SyncPool[T])
	ret.Items = make([]*T, 0, size)
	ret.New = newF
	return ret
}

func SyncPoolGet[T any](p *SyncPool[T]) *T {
	var item *T

	p.M.Lock()
	if len(p.Items) > 0 {
		item = p.Items[len(p.Items)-1]
		p.Items = p.Items[:len(p.Items)-1]
		p.M.Unlock()
	} else {
		p.M.Unlock()
		item = p.New()
	}
	return item
}

func SyncPoolPut[T any](p *SyncPool[T], item *T) {
	p.M.Lock()
	p.Items = append(p.Items, item)
	p.M.Unlock()
}
