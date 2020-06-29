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

package v8worker

/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -Ldepsc++ -ldepsc++ -L/usr/local/lib64/v8 -L/usr/local/Cellar/v8/8.1.307.32/libexec -lstdc++ -lv8 -lv8_libplatform
#include <stdlib.h>
#include "depsc++/v8binding.h"
*/
import "C"
import (
	"errors"
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/common/util"
	"runtime"
)

import "unsafe"
import "sync"

const (
	V8_LIBPATH_MAC   = "/usr/local/Cellar/v8/8.1.307.32/libexec"
	V8_LIBPATH_LINUX = "/usr/local/lib64/v8"
)

var initV8Once sync.Once
var workerTableLock sync.Mutex
var workerTable = make(map[int]*Worker)
var workerTableMaxIndex int = 0

type SendCallback func(worker *Worker, msgType int, msg string, userdata int64)
type RequestCallback func(worker *Worker, msgType int, msg string) string

type msgData struct {
	mtype int
	msg   string
}

type Worker struct {
	cWorker    *C.V8Worker
	tableIndex int
	mutex      *util.Mutex
	sendCb     SendCallback
	requestCb  RequestCallback
	disposed   bool
	running    bool
	msgQueue   []msgData
	expireTime int64
}

//export go_v8SendCallback
func go_v8SendCallback(index C.int, mtype C.int, s *C.char, slen C.int, userdata C.longlong) {
	str := C.GoStringN(s, slen)
	w := workerTableLookup(int(index))
	w.sendCb(w, int(mtype), str, int64(userdata))
}

//export go_v8RequestCallback
func go_v8RequestCallback(index C.int, mtype C.int, s *C.char, slen C.int) *C.char {
	str := C.GoStringN(s, slen)
	w := workerTableLookup(int(index))
	return C.CString(w.requestCb(w, int(mtype), str))
}

func workerTableLookup(index int) *Worker {
	workerTableLock.Lock()
	w := workerTable[index]
	workerTableLock.Unlock()
	return w
}

func Version() string {
	return C.GoString(C.v8_version())
}

func New(sendCb SendCallback, requestCb RequestCallback) *Worker {
	workerTableLock.Lock()
	workerTableMaxIndex++
	w := &Worker{
		sendCb:     sendCb,
		requestCb:  requestCb,
		tableIndex: workerTableMaxIndex,
		mutex:      util.NewMutex(),
	}
	workerTable[w.tableIndex] = w
	workerTableLock.Unlock()

	initV8Once.Do(func() {
		v8Path := V8_LIBPATH_LINUX
		if runtime.GOOS == "darwin" {
			v8Path = V8_LIBPATH_MAC
		}
		icu_path := C.CString(v8Path + "/icudtl.dat")
		C.v8_init(icu_path)
		C.free(unsafe.Pointer(icu_path))
	})

	w.cWorker = C.v8_worker_new(C.int(w.tableIndex))
	return w
}

func (w *Worker) Dispose() {
	if w.disposed {
		return
	}
	w.disposed = true

	workerTableLock.Lock()
	delete(workerTable, w.tableIndex)
	workerTableLock.Unlock()

	C.v8_worker_dispose(w.cWorker)
}

func (w *Worker) Execute(scriptName string, code string) error {
	scriptName_s := C.CString(scriptName)
	code_s := C.CString(code)
	r := C.v8_execute(w.cWorker, scriptName_s, code_s)
	C.free(unsafe.Pointer(scriptName_s))
	C.free(unsafe.Pointer(code_s))

	if r != 0 {
		errStr := C.GoString(C.v8_last_exception(w.cWorker))
		return errors.New(errStr)
	}
	return nil
}

func (w *Worker) Send(mtype int, msg string) error {
	msg_p := C.CString(msg)
	r := C.v8_send(w.cWorker, C.int(mtype), msg_p)
	C.free(unsafe.Pointer(msg_p))
	if r != 0 {
		errStr := C.GoString(C.v8_last_exception(w.cWorker))
		return errors.New(errStr)
	}

	return nil
}

func (w *Worker) SafeSend(mtype int, msg string) error {
	var err error
	w.mutex.Lock()
	if w.running || len(w.msgQueue) > 0 {
		w.msgQueue = append(w.msgQueue, msgData{mtype, msg})
	} else {
		err = w.Send(mtype, msg)
	}
	w.mutex.Unlock()
	return err
}

func (w *Worker) Acquire() bool {
	bOK := false
	if w.mutex.TryLock() {
		if w.running {
			tlog.Error("v8worker still running")
		} else {
			w.running = true
			bOK = true
		}
		w.mutex.Unlock()
	}
	return bOK
}

func (w *Worker) Release() {
	w.mutex.Lock()
	if len(w.msgQueue) > 0 {
		for _, m := range w.msgQueue {
			w.Send(m.mtype, m.msg)
		}
		w.msgQueue = nil
	}
	w.running = false
	w.mutex.Unlock()
}

func (w *Worker) TerminateExecution() {
	C.v8_terminate_execution(w.cWorker)
}

func (w *Worker) SetExpireTime(expireTime int64) {
	w.expireTime = expireTime
}

func (w *Worker) GetExpireTime() int64 {
	return w.expireTime
}
