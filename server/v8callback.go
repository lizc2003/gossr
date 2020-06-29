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

package server

import (
	"encoding/json"
	"fmt"
	"github.com/lizc2003/gossr/common/tlog"
	v8 "github.com/lizc2003/gossr/v8"
	"net/url"
	"strconv"
)

func v8SendCallback(mtype int, msg string, reqId int64) {
	switch mtype {
	case v8.MSGTYPE_SET_URL:
		setTemplateEnv(msg)
	default:
		req := ThisServer.RequstMgr.GetRequest(reqId)
		if req != nil {
			switch mtype {
			case v8.MSGTYPE_SSR_HTML_OK:
				req.bOK = true
				fallthrough
			case v8.MSGTYPE_SSR_HTML_FAIL:
				req.result.Html = msg
				req.wg.Done()
			case v8.MSGTYPE_SSR_CSS:
				req.result.Css = msg
			case v8.MSGTYPE_SSR_META:
				var meta map[string]string
				err := json.Unmarshal([]byte(msg), &meta)
				if err != nil {
					tlog.Error(err)
				} else {
					req.result.Meta = meta
				}
			}
		}
	}
}

func setTemplateEnv(msg string) {
	if len(ThisServer.TemplateUrlEnv) == 0 {
		var dat map[string]string
		err := json.Unmarshal([]byte(msg), &dat)
		var baseUrl string
		var apiUrl string
		if err == nil {
			if v, ok := dat["base"]; ok {
				baseUrl = v
			}
			if v, ok := dat["api"]; ok {
				apiUrl = v
			}
		}
		if baseUrl != "" && apiUrl != "" {
			if ThisServer.IsApiDelegate {
				u, err := url.Parse(apiUrl)
				if err != nil {
					tlog.Error(err)
					return
				}
				u.Host = u.Hostname() + ":" + strconv.FormatInt(int64(ThisServer.HostPort), 10)
				apiUrl = u.String()
			}

			ThisServer.TemplateUrlEnv = fmt.Sprintf(`window.APP_ENV="%s";
			window.BASE_URL="%s";
			window.API_BASE_URL="%s";`, ThisServer.Env, baseUrl, apiUrl)
		}
	}
}
