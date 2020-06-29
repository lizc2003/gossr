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
	"encoding/json"
	"fmt"
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/v8worker"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type xmlHttpReq struct {
	Cmd     string            `json:"cmd"`
	HttpId  int               `json:"httpid"`
	Url     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Post    string            `json:"post"`
	Worker  *v8worker.Worker
	Aborted bool
}

type xmlHttpEvent struct {
	HttpId   int               `json:"httpid"`
	Event    string            `json:"event"`
	Error    string            `json:"error,omitempty"`
	Status   int32             `json:"status,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Response string            `json:"response,omitempty"`
}

func (this *xmlHttpEvent) Reset() {
	this.Event = ""
	this.Error = ""
	this.Status = 0
	this.Headers = nil
	this.Response = ""
}

func processXMLHttpRequestCmd(w *v8worker.Worker, msg string) string {
	var req xmlHttpReq
	err := json.Unmarshal([]byte(msg), &req)
	if err != nil {
		tlog.Error(err)
		return ""
	}
	req.Worker = w

	switch req.Cmd {
	case "open":
		httpid := TheV8Mgr.httpMgr.open(&req)
		return strconv.FormatInt(int64(httpid), 10)
	case "abort":
		TheV8Mgr.httpMgr.abort(req.HttpId)
	}
	return ""
}

func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       60 * time.Second,
			DisableCompression:    true,
			ResponseHeaderTimeout: 6 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 2 * time.Second,
			}).DialContext,
		},
	}
}

type xmlHttpRequestMgr struct {
	mutex           sync.Mutex
	httpClient      *http.Client
	queue           chan *xmlHttpReq
	reqs            map[int]*xmlHttpReq
	internalApiHost string
	internalApiIp   string
	internalApiPort int32
	maxId           int
}

func NewXmlHttpRequestMgr(maxCount int, internalApiHost string, internalApiIp string, internalApiPort int32) *xmlHttpRequestMgr {
	queue := make(chan *xmlHttpReq, maxCount*2)
	reqs := make(map[int]*xmlHttpReq)
	that := &xmlHttpRequestMgr{httpClient: newHttpClient(), queue: queue, reqs: reqs,
		internalApiHost: internalApiHost, internalApiIp: internalApiIp, internalApiPort: internalApiPort}

	for i := 0; i < maxCount; i++ {
		go func() {
			for req := range queue {
				that.performRequest(req)

				that.mutex.Lock()
				delete(that.reqs, req.HttpId)
				that.mutex.Unlock()
			}
		}()
	}
	return that
}

func (this *xmlHttpRequestMgr) open(req *xmlHttpReq) int {
	if len(this.internalApiIp) > 0 {
		pos := strings.Index(req.Url, "://")
		if pos > 0 {
			req.Url = req.Url[pos+3:]
			pos = strings.Index(req.Url, "/")
			req.Url = req.Url[pos:]
		}
		req.Url = fmt.Sprintf("http://%s:%d%s", this.internalApiIp, this.internalApiPort, req.Url)
	}

	this.mutex.Lock()
	this.maxId++
	req.HttpId = this.maxId
	this.reqs[req.HttpId] = req
	this.mutex.Unlock()

	beginTime := time.Now()
	this.queue <- req
	tlog.Infof("xhr request %d: %s, wait time: %v", req.HttpId, req.Url, time.Since(beginTime))
	return req.HttpId
}

func (this *xmlHttpRequestMgr) abort(httpId int) {
	this.mutex.Lock()
	if req, ok := this.reqs[httpId]; ok {
		req.Aborted = true
	}
	this.mutex.Unlock()
}

func (this *xmlHttpRequestMgr) performRequest(req *xmlHttpReq) {
	worker := req.Worker
	evt := xmlHttpEvent{HttpId: req.HttpId}
	if req.Aborted {
		evt.Event = "onfinish"
		sendHttpEvent(worker, &evt)
		return
	}

	evt.Event = "onstart"
	sendHttpEvent(worker, &evt)

	url := req.Url
	var request *http.Request
	var err error
	if len(req.Post) == 0 {
		request, err = http.NewRequest(req.Method, url, nil)
	} else {
		request, err = http.NewRequest(req.Method, url, strings.NewReader(req.Post))
		if _, ok := req.Headers["Content-Type"]; !ok {
			c := req.Post[:1]
			if c == "{" || c == "[" {
				req.Headers["Content-Type"] = "application/json;charset=UTF-8"
			} else {
				req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
			}
		}
	}
	if err != nil {
		sendHttpErrorEvent(worker, &evt, err)
		return
	}
	if len(this.internalApiHost) > 0 {
		request.Host = this.internalApiHost
	}
	for k, v := range req.Headers {
		if k == "SSR-Ctx" {
			if v != "" {
				var headers map[string]string
				err := json.Unmarshal([]byte(v), &headers)
				if err == nil {
					for kk, vv := range headers {
						if vv != "" {
							kk = strings.ReplaceAll(kk, "_", "-")
							tlog.Infof("xhr header %s: %s", kk, vv)
							request.Header.Set(kk, vv)
						}
					}
				}
			}
		} else {
			request.Header.Set(k, v)
		}
	}
	if req.Aborted {
		evt.Event = "onfinish"
		sendHttpEvent(worker, &evt)
		return
	}

	resp, err := this.httpClient.Do(request)
	if req.Aborted {
		evt.Event = "onfinish"
		sendHttpEvent(worker, &evt)
		return
	}
	if err != nil {
		sendHttpErrorEvent(worker, &evt, err)
		return
	}
	if resp == nil {
		err = fmt.Errorf("resp is nil: %s", url)
		sendHttpErrorEvent(worker, &evt, err)
		return
	}

	evt.Event = "onheader"
	evt.Status = int32(resp.StatusCode)
	evt.Headers = make(map[string]string)
	for k, v := range resp.Header {
		evt.Headers[k] = strings.Join(v, "&")
	}
	sendHttpEvent(worker, &evt)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if req.Aborted {
		evt.Event = "onfinish"
		sendHttpEvent(worker, &evt)
		return
	}
	if err != nil {
		sendHttpErrorEvent(worker, &evt, err)
		return
	}

	evt.Event = "onend"
	evt.Response = string(body)
	sendHttpEvent(worker, &evt)

	evt.Event = "onfinish"
	sendHttpEvent(worker, &evt)
}

func sendHttpErrorEvent(w *v8worker.Worker, evt *xmlHttpEvent, err error) {
	tlog.Error(err)

	evt.Event = "onerror"
	evt.Error = err.Error()
	sendHttpEvent(w, evt)

	evt.Event = "onfinish"
	sendHttpEvent(w, evt)
}

func sendHttpEvent(w *v8worker.Worker, evt *xmlHttpEvent) {
	s, _ := json.Marshal(evt)
	err := w.SafeSend(MSGTYPE_HTTP_CALLBACK, string(s))
	if err != nil {
		tlog.Error(err)
	}
	evt.Reset()
}
