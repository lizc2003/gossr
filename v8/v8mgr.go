package v8

import (
	"fmt"
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/v8worker"
	"sync/atomic"
	"time"
)

const (
	DELETE_DALAY_TIME = 2 * time.Minute
)

type V8MgrConfig struct {
	Env             string
	JsPaths         []string
	MaxWorkerCount  int32
	WorkerLifeTime  int
	InternalApiHost string
	InternalApiIp   string
	InternalApiPort int32
}

type V8Mgr struct {
	env                string
	httpMgr            *xmlHttpRequestMgr
	workers            chan *v8worker.Worker
	workerLifeTime     int64
	maxWorkerCount     int32
	currentWorkerCount int32
}

var TheV8Mgr *V8Mgr

func NewV8Mgr(c *V8MgrConfig) (*V8Mgr, error) {
	initV8Module(c.JsPaths)
	initV8NewJs()

	worker, err := newV8Worker(c.Env)
	if err != nil {
		return nil, err
	}

	lifeTime := int64(c.WorkerLifeTime)
	worker.SetExpireTime(time.Now().Unix() + lifeTime)
	workers := make(chan *v8worker.Worker, c.MaxWorkerCount+100)
	workers <- worker

	TheV8Mgr = &V8Mgr{env: c.Env,
		httpMgr: NewXmlHttpRequestMgr(int(c.MaxWorkerCount)*2, c.InternalApiHost, c.InternalApiIp, c.InternalApiPort),
		workers: workers, workerLifeTime: lifeTime,
		maxWorkerCount: c.MaxWorkerCount, currentWorkerCount: 1}
	return TheV8Mgr, nil
}

func (this *V8Mgr) Execute(name string, code string) error {
	w := this.acquireWorker()
	err := w.Execute(name, code)
	if err != nil {
		tlog.Error(err)
	}
	this.releaseWorker(w)
	return err
}

func (this *V8Mgr) GetInternelApiUrl() string {
	if this.httpMgr.internalApiHost != "" {
		return fmt.Sprintf("http://%s:%d", this.httpMgr.internalApiHost, this.httpMgr.internalApiPort)
	}
	return ""
}

func (this *V8Mgr) acquireWorker() *v8worker.Worker {
	var busyWorkers []*v8worker.Worker
	for {
		var ret *v8worker.Worker
		bEmpty := false
		select {
		case worker := <-this.workers:
			if worker.Acquire() {
				ret = worker
			} else {
				busyWorkers = append(busyWorkers, worker)
			}
		default:
			if this.currentWorkerCount < this.maxWorkerCount {
				worker, err := newV8Worker(this.env)
				if err == nil {
					atomic.AddInt32(&this.currentWorkerCount, 1)
					worker.SetExpireTime(time.Now().Unix() + this.workerLifeTime)
					worker.Acquire()
					ret = worker
				} else {
					bEmpty = true
				}
			} else {
				bEmpty = true
			}
		}

		if ret != nil {
			for _, w := range busyWorkers {
				this.workers <- w
			}
			return ret
		} else if bEmpty {
			if len(busyWorkers) > 0 {
				for _, w := range busyWorkers {
					this.workers <- w
				}
				busyWorkers = busyWorkers[:0]
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (this *V8Mgr) releaseWorker(worker *v8worker.Worker) {
	if worker != nil {
		worker.Release()

		if time.Now().Unix() >= worker.GetExpireTime() {
			atomic.AddInt32(&this.currentWorkerCount, -1)

			go func(w *v8worker.Worker) {
				time.Sleep(DELETE_DALAY_TIME)
				w.Dispose()
			}(worker)
		} else {
			this.workers <- worker
		}
	}
}
