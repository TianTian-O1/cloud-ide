package waitgroup

import "sync"

type WaitGroupWapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWapper) Go(fun func()) {
	w.Add(1)
	go func() {
		fun()
		w.Done()
	}()
}
