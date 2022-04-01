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
	"os"
	"strconv"
	"strings"
	"time"
)

type RobotFeishu struct {
	env        string
	url        string
	secret     string
	httpClient *http.Client
}

// https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN
func NewRobotFeishu(env string, app string, url string, secret string) *RobotFeishu {
	host, _ := os.Hostname()
	r := &RobotFeishu{
		env:        "env: " + env + "\n" + "host: " + host + "\n" + "app: " + app + "\n",
		url:        url,
		secret:     secret,
		httpClient: NewHttpClient(false),
	}
	return r
}

func (this *RobotFeishu) SendMsg(msg string) string {
	var timestamp string
	var sign string
	if this.secret != "" {
		timestamp = strconv.FormatInt(time.Now().UnixMilli(), 10)
		strSign := timestamp + "\n" + this.secret

		h := hmac.New(sha256.New, []byte(strSign))
		h.Write([]byte{})
		sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
	}

	type msgText struct {
		Text string `json:"text"`
	}

	type msgData struct {
		Timestamp string  `json:"timestamp,omitempty"`
		Sign      string  `json:"sign,omitempty"`
		Msgtype   string  `json:"msg_type"`
		Content   msgText `json:"content"`
	}

	b := strings.Builder{}
	b.WriteString(this.env)
	b.WriteString("time: ")
	b.WriteString(FormatFullTime(time.Now()))
	b.WriteByte('\n')
	b.WriteString(msg)

	data := msgData{Timestamp: timestamp, Sign: sign,
		Msgtype: "text", Content: msgText{Text: b.String()}}

	var ret string
	err := HttpPost(this.httpClient, this.url, nil, data, &ret)
	if err != nil {
		tlog.Error(err)
	}
	return ret
}
