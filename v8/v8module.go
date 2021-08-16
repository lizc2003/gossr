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

package v8

import (
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
		var firstErr error
		for _, path := range gJsPaths {
			filename := path + jsId
			content, err = ioutil.ReadFile(filename)
			if err == nil {
				m.Filename = filename
				break
			} else if firstErr == nil {
				firstErr = err
			}
		}
		if err != nil {
			err = firstErr
		}
	}

	if err != nil {
		tlog.Error(err)
		m.Err = err
		return
	}

	m.Dirname = path.Dir(m.Filename)

	var b strings.Builder
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
