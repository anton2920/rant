package main

type SyncQueue[T any] struct {
	M     Mutex
	Items []*T
	Pos   int
}

func SyncQueueGet[T any](q *SyncQueue[T]) *T {
	var ret *T

	q.M.Lock()
	if len(q.Items) > q.Pos {
		ret = q.Items[q.Pos]
		q.Pos++
	} else {
		q.Pos = 0
		q.Items = q.Items[:0]
	}
	q.M.Unlock()

	return ret
}

func SyncQueuePut[T any](q *SyncQueue[T], item *T) {
	q.M.Lock()
	q.Items = append(q.Items, item)
	q.M.Unlock()
}
