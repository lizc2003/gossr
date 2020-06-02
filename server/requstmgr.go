package server

import (
	"sync"
)

type SsrContext struct {
	Title       string `json:"title"`
	Keywords    string `json:"keywords"`
	Description string `json:"description"`
	Ogimage     string `json:"ogimage"`
	Canolink    string `json:"canolink"`
	State       string `json:"state"`
	Initscript  string `json:"initscript"`
	Seocontent  string `json:"seocontent"`
	Schema      string `json:"schema"`
	Metaheader  string `json:"metaheader"`
}

type SsrResult struct {
	Html string     `json:"html"`
	Css  string     `json:"css"`
	Ctx  SsrContext `json:"ctx"`
}

type Request struct {
	wg     sync.WaitGroup
	reqId  int64
	result SsrResult
	bOK    bool
}

type RequestMgr struct {
	mutex    sync.Mutex
	requests map[int64]*Request
	maxId    int64
}

func NewRequestMgr() *RequestMgr {
	return &RequestMgr{requests: make(map[int64]*Request)}
}

func (this *RequestMgr) NewRequest() *Request {
	req := &Request{}
	req.wg.Add(1)

	this.mutex.Lock()
	this.maxId++
	req.reqId = this.maxId
	this.requests[req.reqId] = req
	this.mutex.Unlock()
	return req
}

func (this *RequestMgr) DestroyRequest(reqId int64) {
	this.mutex.Lock()
	if _, ok := this.requests[reqId]; ok {
		delete(this.requests, reqId)
	}
	this.mutex.Unlock()
}

func (this *RequestMgr) GetRequest(reqId int64) *Request {
	var req *Request
	this.mutex.Lock()
	if v, ok := this.requests[reqId]; ok {
		req = v
	}
	this.mutex.Unlock()
	return req
}
