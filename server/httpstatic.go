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
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type DebugTransport struct{}

func (DebugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(r, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(r)
}

func GetStaticAndProxyHandler(urlPrefix, rootPath string) gin.HandlerFunc {
	fileServer := http.FileServer(gin.Dir(rootPath, false))
	fileServer = http.StripPrefix(urlPrefix, fileServer)

	var proxyServer *httputil.ReverseProxy
	if ThisServer.IsApiDelegate {
		apiurl := ThisServer.V8Mgr.GetInternelApiUrl()
		if apiurl != "" {
			proxyUrl, _ := url.Parse(apiurl)
			proxyServer = httputil.NewSingleHostReverseProxy(proxyUrl)
			targetHost := proxyUrl.Host
			originD := proxyServer.Director
			proxyServer.Director = func(r *http.Request) {
				originD(r)          // call default director
				r.Host = targetHost // set Host header as expected by target
			}
			//proxyServer.Transport = DebugTransport{}
		}
	}

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, urlPrefix) {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		if proxyServer != nil {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				proxyServer.ServeHTTP(c.Writer, c.Request)
				c.Abort()
				return
			}
		}
	}
}
