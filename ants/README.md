repo: https://github.com/panjf2000/ants

### 1. worker 是怎么执行的
- `retrieveWorker`尝试从workerQueue获取worker，获取不到且没有达到数量限制时新创建worker。 如果worker都在运行中，且pool为非阻塞模式或当前阻塞数量大于阈值的时候返回overload错误。

- 在初始化worker之后， 从channel中读取写入的function, 执行完后worker再放入worker队列中。 在这个worker的run函数中启动的goroutine中获取channel中的function来处理，达到了goroutine的复用。
```go
func (w *goWorker) run() {
	w.pool.addRunning(1)
	go func() {
		defer func() {
			if w.pool.addRunning(-1) == 0 && w.pool.IsClosed() {
				w.pool.once.Do(func() {
					close(w.pool.allDone)
				})
			}
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				if ph := w.pool.options.PanicHandler; ph != nil {
					ph(p)
				} else {
					w.pool.options.Logger.Printf("worker exits from panic: %v\n%s\n", p, debug.Stack())
				}
			}
			// Call Signal() here in case there are goroutines waiting for available workers.
			w.pool.cond.Signal()
		}()

		for f := range w.task {
			if f == nil {
				return
			}
			f()
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}
```

### 2. 新的函数要被执行时如何阻塞和唤醒的
- 在`retrieveWorker`中通过`sync.Cond`等待，
```
	// Otherwise, we'll have to keep them blocked and wait for at least one worker to be put back into pool.
	p.addWaiting(1)
	p.cond.Wait() // block and wait for an available worker
	p.addWaiting(-1)
```
- 执行task之后，通过`revertWorker`函数将worker放入队列，并通知。
```go
// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *Pool) revertWorker(worker *goWorker) bool {
	if capacity := p.Cap(); (capacity > 0 && p.Running() > capacity) || p.IsClosed() {
		p.cond.Broadcast()
		return false
	}

	worker.lastUsed = p.nowTime()

	p.lock.Lock()
	// To avoid memory leaks, add a double check in the lock scope.
	// Issue: https://github.com/panjf2000/ants/issues/113
	if p.IsClosed() {
		p.lock.Unlock()
		return false
	}
	if err := p.workers.insert(worker); err != nil {
		p.lock.Unlock()
		return false
	}
	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()

	return true
}
```