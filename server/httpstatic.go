package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func GetStaticAndProxyHandler(urlPrefix, rootPath string) gin.HandlerFunc {
	fileServer := http.FileServer(gin.Dir(rootPath, false))
	fileServer = http.StripPrefix(urlPrefix, fileServer)

	var proxyServer *httputil.ReverseProxy
	if ThisServer.IsDevEnv {
		apiurl := ThisServer.V8Mgr.GetInternelApiUrl()
		if apiurl != "" {
			proxyUrl, _ := url.Parse(apiurl)
			proxyServer = httputil.NewSingleHostReverseProxy(proxyUrl)
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
