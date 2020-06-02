package server

import (
	"encoding/json"
	"fmt"
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
	ssrHeaders := map[string]string{"Cookie": cookie}
	for _, k := range ThisServer.SsrHeaders {
		v := c.GetHeader(k)
		if v == "" && k == "X-Forwarded-For" {
			v = c.ClientIP()
		}
		ssrHeaders[strings.ReplaceAll(k, "-", "_")] = v
	}

	tlog.Infof("%s", url)
	result, bOK := generateSsrResult(url, ssrHeaders)

	if !bOK && ThisServer.RedirectOnerror != "" {
		tlog.Warningf("redirect: %s?%s", reqURL.Path, reqURL.RawQuery)
		result.Ctx.Initscript = fmt.Sprintf("window.location.href = '%s';",
			ThisServer.RedirectOnerror)
	}

	templateName := ThisServer.SsrTemplate
	c.HTML(http.StatusOK, templateName, gin.H{
		"html":        template.HTML(result.Html),
		"styles":      template.HTML(result.Css),
		"title":       result.Ctx.Title,
		"keywords":    result.Ctx.Keywords,
		"description": result.Ctx.Description,
		"metaheader":  template.HTML(result.Ctx.Metaheader),
		"ogimage":     template.HTML(result.Ctx.Ogimage),
		"canolink":    result.Ctx.Canolink,
		"state":       template.JS(result.Ctx.State),
		"initscript":  template.JS(result.Ctx.Initscript),
		"schema":      template.JS(result.Ctx.Schema),
		"seocontent":  template.HTML(result.Ctx.Seocontent),
		"appenv":      ThisServer.Env,
		"baseurl":     ThisServer.tmplateBaseUrl,
		"ajaxbaseurl": ThisServer.tmplateAjaxBaseUrl,
	})
}

func generateSsrResult(url string, ssrHeaders map[string]string) (SsrResult, bool) {
	req := ThisServer.RequstMgr.NewRequest()

	headerJson, _ := json.Marshal(ssrHeaders)
	var jsCode strings.Builder
	jsCode.Grow(renderJsLength + len(headerJson) + len(url) + 30)
	jsCode.WriteString(renderJsPart1)
	jsCode.WriteString(`{v8reqid:`)
	jsCode.WriteString(strconv.FormatInt(req.reqId, 10))
	jsCode.WriteString(`,cookie:"a=b",url:"`)
	jsCode.WriteString(url)
	jsCode.WriteString(`",ssrHeaders:`)
	jsCode.Write(headerJson)
	jsCode.WriteString(`}`)
	jsCode.WriteString(renderJsPart2)

	fmt.Println(jsCode.String())

	err := ThisServer.V8Mgr.Execute("bundle.js", jsCode.String())

	if err == nil {
		req.wg.Wait()
	} else {
		req.result.Html = err.Error()
	}
	ThisServer.RequstMgr.DestroyRequest(req.reqId)

	return req.result, req.bOK
}

func generateUUID() string {
	uuid := uuid.NewV4().String()
	return strings.Replace(uuid, "-", "", -1)
}
