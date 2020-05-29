package v8

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/v8worker"
	"io/ioutil"
	"path"
	"strings"
)

var gJsPaths []string

func initV8Module(jsPaths []string) {
	gJsPaths = jsPaths
}

type v8module struct {
	Err      error  `json:"err"`
	Source   string `json:"source"`
	Id       string `json:"id"`
	Filename string `json:"filename"`
	Dirname  string `json:"dirname"`
	isMain   bool
}

func (m *v8module) load() {
	tlog.Infof("require: %s", m.Id)

	jsId := m.Id
	if !strings.HasSuffix(jsId, ".js") {
		jsId += ".js"
	}

	var content []byte
	var err error
	if len(gJsPaths) == 0 {
		err = errors.New("js paths not inited")
	} else {
		for _, path := range gJsPaths {
			filename := path + jsId
			content, err = ioutil.ReadFile(filename)
			if err == nil {
				m.Filename = filename
				break
			}
		}
	}

	if err != nil {
		tlog.Error(err)
		m.Err = err
		return
	}

	m.Dirname = path.Dir(m.Filename)

	var b bytes.Buffer
	if m.isMain {
		b.WriteString(fmt.Sprintf(
			"var main = new NativeModule({ id: '%s', filename: '%s', dirname: '%s' });\n",
			m.Id, m.Filename, m.Dirname))
	}

	b.WriteString("(function (exports, require, module, __filename, __dirname) { ")
	if m.isMain {
		b.WriteString("\nrequire.main = module;")
	}
	b.Write(content)

	if m.isMain {
		b.WriteString("\n}")
		b.WriteString("(main.exports, NativeModule.require, main, main.filename, main.dirname));")
	} else {
		b.WriteString("\n});")
	}
	m.Source = b.String()
}

const nativeModuleJsContent = `
	'use strict';
	function NativeModule(rawModule) {
		this.filename = rawModule.filename;
		this.dirname = rawModule.dirname;
		this.id = rawModule.id;
		this.exports = {};
		this.loaded = false;
		this._source = rawModule.source;
	}
	NativeModule.require = function(id) {
		var source = v8worker.request(10, id);
		var rawModule = JSON.parse(source);
		if (rawModule.err) {
			throw new RangeError(JSON.stringify(rawModule.err));
		}
		var nativeModule = new NativeModule(rawModule);
		nativeModule.compile();
		return nativeModule.exports;
	};
	NativeModule.prototype.compile = function() {
		var fn = eval(this._source);
		fn(this.exports, NativeModule.require, this, this.filename, this.dirname);
		this.loaded = true;
	};
`

func loadMainModule(w *v8worker.Worker, id string) error {
	m := v8module{Id: id, isMain: true}
	m.load()
	if m.Err != nil {
		return m.Err
	}
	return w.Execute(id, m.Source)
}

func requireModule(id string) string {
	m := v8module{Id: id}
	m.load()
	bytes, _ := json.Marshal(m)
	return string(bytes)
}
