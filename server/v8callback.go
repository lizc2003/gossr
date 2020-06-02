package server

import (
	"encoding/json"
	"github.com/lizc2003/gossr/common/tlog"
	v8 "github.com/lizc2003/gossr/v8"
)

func v8SendCallback(mtype int, msg string, reqId int64) {
	switch mtype {
	case v8.MSGTYPE_SET_BASEURL:
		if len(ThisServer.tmplateBaseUrl) == 0 {
			ThisServer.tmplateBaseUrl = msg
		}
	case v8.MSGTYPE_SET_AJAXBASEURL:
		if len(ThisServer.tmplateAjaxBaseUrl) == 0 {
			ThisServer.tmplateAjaxBaseUrl = msg
		}
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
			case v8.MSGTYPE_SSR_CONTEXT:
				var ctx SsrContext
				err := json.Unmarshal([]byte(msg), &ctx)
				if err != nil {
					tlog.Error(err)
				} else {
					req.result.Ctx = ctx
				}
			}
		}
	}
}
