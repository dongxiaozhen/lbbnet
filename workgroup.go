package lbbnet

import (
	"runtime"
	"sync"

	log "github.com/donnie4w/go-logger/logger"
)

type WorkTask struct {
	pipeLen  uint64
	cacheLen int
	process  []chan func()
	sync.WaitGroup
	index uint64
	mu    sync.Mutex
}

func NewWorkTask(plen, clen int) *WorkTask {
	log.Debug("NewWorkTask", plen, clen)
	return &WorkTask{pipeLen: uint64(plen), cacheLen: clen, process: make([]chan func(), plen)}
}

func (t *WorkTask) SendTask(f func()) {
	t.mu.Lock()
	a := t.index % t.pipeLen
	t.index++
	t.mu.Unlock()

	t.process[a] <- f
}

func (t *WorkTask) Run() {
	for i := uint64(0); i < t.pipeLen; i++ {
		t.process[i] = make(chan func(), t.cacheLen)
		t.Add(1)
		go t.hander(i)
	}
}

func (t *WorkTask) hander(index uint64) {
	defer t.Done()

	ch := t.process[index]
	for ff := range ch {
		f := func() {
			defer CapturePanic()
			ff()
		}
		f()
	}
}

func (t *WorkTask) Stop() {
	for _, c := range t.process {
		close(c)
	}
	t.Wait()
}

func CapturePanic() {
	if err := recover(); err != nil {
		buf := make([]byte, 1<<20)
		runtime.Stack(buf, true)
		log.Debug(buf)
	}
}
