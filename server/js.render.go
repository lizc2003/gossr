package server

import (
	"strings"
)

const renderJsContent = `
serverBundle(RENDER_CONTEXT).then((appObj) => {
	var app = appObj.app
	var context = appObj.context
	try {
		var contextData = {
			title: context.title,
			keywords: context.keywords,
			description: context.description,
			ogimage: context.ogimage,
			canolink: context.canolink,
			initscript: context.initscript,
			seocontent: context.seocontent,
			schema: context.schema,
			metaheader: context.metaheader,
			state: JSON.stringify(context.state)
		};
		v8worker.send(83, JSON.stringify(contextData), context.v8reqid)

		renderToString(app, context, (err, html) => {
			try {
				if (err) {
					if (err.code == 404) {
						v8worker.send(81, "404 Page not found", context.v8reqid)
					} else if (err.code) {
						v8worker.send(81, err.code + " Internal Server Error", context.v8reqid)
					} else {
						console.error(err)
						v8worker.send(81, err, context.v8reqid)
					}
				} else {
					if (context.styles) {
						v8worker.send(82, context.styles, context.v8reqid)
					}
					v8worker.send(80, html, context.v8reqid)
				}
			} catch(e) {
				console.error(e);
				v8worker.send(81, e, context.v8reqid)
			}
		})
	} catch(e) {
		console.error(e);
		v8worker.send(81, e, context.v8reqid)
	}
}, (errObj) => {
	var err = errObj.err
	var context = errObj.context
	if (err.code == 404) {
		v8worker.send(81, "404 Page not found", context.v8reqid)
	} else {
		console.error(err)
		v8worker.send(81, err, context.v8reqid)
	}
});
`

var renderJsPart1 string
var renderJsPart2 string
var renderJsLength int

func init() {
	s := "RENDER_CONTEXT"
	idx := strings.Index(renderJsContent, s)
	renderJsPart1 = renderJsContent[:idx]
	renderJsPart2 = renderJsContent[idx+len(s):]
	renderJsLength = len(renderJsPart1) + len(renderJsPart2)
}
