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
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/common/util"
	uuid "github.com/satori/go.uuid"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleSsrRequest(c *gin.Context) {
	reqURL := c.Request.URL
	url := reqURL.Path
	if len(reqURL.RawQuery) > 0 {
		url += "?"
		url += reqURL.RawQuery
	}

	cookie := c.GetHeader("Cookie")
	if ThisServer.ClientCookie != "" {
		var clientId string
		cookieName := ThisServer.ClientCookie
		cookieVal, err := c.Request.Cookie(cookieName)
		if err == nil && len(cookieVal.Value) > 0 {
			clientId = cookieVal.Value
		} else {
			clientId = generateUUID() + strconv.FormatInt(int64(rand.Int31n(10)), 10)
			c.SetCookie(cookieName, clientId, 24*3600*365*10,
				"/", util.GetDomainFromHost(c.Request.Host), false, false)
			if len(cookie) > 0 {
				cookie = cookieName + "=" + clientId + "; " + cookie
			} else {
				cookie = cookieName + "=" + clientId
			}
		}
	}
	ssrCtx := map[string]string{"Cookie": cookie}
	for _, k := range ThisServer.SsrCtx {
		v := c.GetHeader(k)
		if v == "" && k == "X-Forwarded-For" {
			v = c.ClientIP()
		}
		ssrCtx[strings.ReplaceAll(k, "-", "_")] = v
	}

	tlog.Infof("http request: %s", url)
	result, bOK, bNoV8 := generateSsrResult(url, ssrCtx)

	if !bOK && !bNoV8 && ThisServer.RedirectOnerror != "" && reqURL.Path != ThisServer.RedirectOnerror {
		tlog.Errorf("redirect: %s?%s", reqURL.Path, reqURL.RawQuery)
		c.Redirect(302, ThisServer.RedirectOnerror)
		return
	}

	outputHtml(c, result)
}

func outputHtml(c *gin.Context, result SsrResult) {
	templName := ThisServer.SsrTemplate
	templObj := gin.H{
		"Html":   template.HTML(result.Html),
		"Css":    template.HTML(result.Css),
		"UrlEnv": template.JS(ThisServer.TemplateUrlEnv),
	}
	for k, v := range result.Meta {
		if v != "" {
			ktype := ThisServer.TemplateVars[k]
			switch ktype {
			case "js":
				templObj[k] = template.JS(v)
			case "html":
				templObj[k] = template.HTML(v)
			default:
				templObj[k] = v
			}
		}
	}
	c.HTML(http.StatusOK, templName, templObj)
}

func generateSsrResult(url string, ssrCtx map[string]string) (SsrResult, bool, bool) {
	req := ThisServer.RequstMgr.NewRequest()

	ssrCtxJson, _ := json.Marshal(ssrCtx)
	urlJson, _ := json.Marshal(url)

	var jsCode strings.Builder
	jsCode.Grow(renderJsLength + len(ssrCtxJson) + len(urlJson) + 28)
	jsCode.WriteString(renderJsPart1)
	jsCode.WriteString(`{v8reqId:`)
	jsCode.WriteString(strconv.FormatInt(req.reqId, 10))
	jsCode.WriteString(`,url:`)
	jsCode.Write(urlJson)
	jsCode.WriteString(`,ssrCtx:`)
	jsCode.Write(ssrCtxJson)
	jsCode.WriteString(`}`)
	jsCode.WriteString(renderJsPart2)

	//fmt.Println(jsCode.String())

	err, bNoV8 := ThisServer.V8Mgr.Execute("bundle.js", jsCode.String())

	if err == nil {
		req.wg.Wait()
	} else {
		req.result.Html = err.Error()
	}
	ThisServer.RequstMgr.DestroyRequest(req.reqId)

	return req.result, req.bOK, bNoV8
}

func generateUUID() string {
	uuid := uuid.NewV4().String()
	return strings.Replace(uuid, "-", "", -1)
}
