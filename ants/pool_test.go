package ants

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
)

// 测试pool并发的增加变量的值， 观察结果是否和期望的一致
func TestPoolSubmit(t *testing.T) {
	pool, _ := ants.NewPool(100)
	defer pool.Release()

	var val atomic.Int32
	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		pool.Submit(func() {
			val.Add(1)
			wg.Done()
		})
	}

	wg.Wait()
	if val.Load() != 100 {
		t.Errorf("expected 100, got:%d", val.Load())
	}

}

// 测试pool并发执行任务时，是否会阻塞
func TestPoolBlock(t *testing.T) {
	pool, _ := ants.NewPool(1, ants.WithOptions(ants.Options{
		Nonblocking: true,
	}))
	defer pool.Release()

	errCh := make(chan error)
	defer close(errCh)

	err := pool.Submit(func() {
		time.Sleep(time.Second * 3)
	})
	if err != nil {
		t.Errorf("expected err is nil, got:%v", err)
	}

	// make sure the second goroutine start after the previous one
	err = pool.Submit(func() {
		time.Sleep(time.Second * 1)
	})

	if err == nil || err != ants.ErrPoolOverload {
		t.Errorf("expected pool overload, got:%v", err)
	}

}
