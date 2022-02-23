// Copyright 2020-present, lizc2003@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v8

import (
	"errors"
	"fmt"
	"github.com/lizc2003/gossr/alarm"
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/v8worker"
	"os"
	"sync/atomic"
	"time"
)

const (
	DELETE_DALAY_TIME = 2 * time.Minute
	V8_REQ_TIMEOUT    = 8 // seconds
	V8_EXIT_THRESHOLD = 1000
)

type V8SendCallback func(msgType int, msg string, userdata int64)

type V8MgrConfig struct {
	Env             string
	JsPaths         []string
	MaxWorkerCount  int32
	WorkerLifeTime  int
	InternalApiHost string
	InternalApiIp   string
	InternalApiPort int32
	SendCallback    V8SendCallback
}

type V8Mgr struct {
	env                string
	httpMgr            *xmlHttpRequestMgr
	SendCallback       V8SendCallback
	workers            chan *v8worker.Worker
	workerLifeTime     int64
	maxWorkerCount     int32
	currentWorkerCount int32
	unavailableCount   int32
}

var TheV8Mgr *V8Mgr

func NewV8Mgr(c *V8MgrConfig) (*V8Mgr, error) {
	initV8Module(c.JsPaths)
	initV8NewJs()

	TheV8Mgr = &V8Mgr{env: c.Env,
		httpMgr:        NewXmlHttpRequestMgr(int(c.MaxWorkerCount)*2, c.InternalApiHost, c.InternalApiIp, c.InternalApiPort),
		SendCallback:   c.SendCallback,
		workerLifeTime: int64(c.WorkerLifeTime),
		maxWorkerCount: c.MaxWorkerCount}

	worker, err := newV8Worker(c.Env)
	if err != nil {
		return nil, err
	}

	worker.SetExpireTime(time.Now().Unix() + int64(c.WorkerLifeTime))
	workers := make(chan *v8worker.Worker, c.MaxWorkerCount+100)
	workers <- worker

	TheV8Mgr.workers = workers
	TheV8Mgr.currentWorkerCount = 1
	return TheV8Mgr, nil
}

func (this *V8Mgr) Execute(name string, code string) (error, bool) {
	w := this.acquireWorker()
	if w == nil {
		err := errors.New("V8 worker not available.")
		tlog.Error(err)
		alarm.SendMessage(err.Error())
		return err, true
	}
	err := w.Execute(name, code)
	if err != nil {
		tlog.Error(err)
	}
	this.releaseWorker(w)
	return err, false
}

func (this *V8Mgr) GetInternelApiUrl() string {
	if this.httpMgr.internalApiHost != "" {
		return fmt.Sprintf("http://%s:%d", this.httpMgr.internalApiHost, this.httpMgr.internalApiPort)
	}
	return ""
}

func (this *V8Mgr) acquireWorker() *v8worker.Worker {
	var busyWorkers []*v8worker.Worker
	reqStartTime := time.Now().Unix()
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
				atomic.AddInt32(&this.currentWorkerCount, 1)
				worker, err := newV8Worker(this.env)
				if err == nil {
					worker.SetExpireTime(time.Now().Unix() + this.workerLifeTime)
					worker.Acquire()
					ret = worker
				} else {
					atomic.AddInt32(&this.currentWorkerCount, -1)
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

		if time.Now().Unix()-reqStartTime > V8_REQ_TIMEOUT {
			errCount := atomic.AddInt32(&this.unavailableCount, 1)
			if errCount == V8_EXIT_THRESHOLD {
				errMsg := "v8 unavailable too many times, exit!"
				tlog.Error(errMsg)
				alarm.SendMessage(errMsg)
				os.Exit(1)
			}
			return nil
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
