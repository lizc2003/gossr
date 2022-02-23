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

package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/lizc2003/gossr/common/tlog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type RobotDingDing struct {
	env        string
	url        string
	secret     string
	httpClient *http.Client
}

// https://developers.dingtalk.com/document/app/custom-robot-access
func NewRobotDingDing(env string, app string, url string, secret string) *RobotDingDing {
	host, _ := os.Hostname()
	r := &RobotDingDing{
		env:        "env: " + env + "\n" + "host: " + host + "\n" + "app: " + app + "\n",
		url:        url,
		secret:     secret,
		httpClient: NewHttpClient(false),
	}
	return r
}

func (this *RobotDingDing) SendMsg(msg string) string {
	var ddurl string
	if this.secret != "" {
		timestamp := strconv.FormatInt(GetMilliUnixTime(), 10)
		strSign := timestamp + "\n" + this.secret

		h := hmac.New(sha256.New, []byte(this.secret))
		h.Write([]byte(strSign))
		sign := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))
		ddurl = this.url + "&timestamp=" + timestamp + "&sign=" + sign
	} else {
		ddurl = this.url
	}

	type msgText struct {
		Content string `json:"content"`
	}
	type msgData struct {
		Msgtype string  `json:"msgtype"`
		Text    msgText `json:"text"`
	}

	b := strings.Builder{}
	b.WriteString(this.env)
	b.WriteString("time: ")
	b.WriteString(FormatFullTime(time.Now()))
	b.WriteByte('\n')
	b.WriteString(msg)

	data := msgData{Msgtype: "text", Text: msgText{Content: b.String()}}
	var ret string
	err := HttpPost(this.httpClient, ddurl, nil, data, &ret)
	if err != nil {
		tlog.Error(err)
	}
	return ret
}
