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

const globalJsContent = `
this["console"] = {
	debug(...args) {
		v8worker.print(0, ...args)
	},
	log(...args) {
		v8worker.print(1, ...args)
	},
	info(...args) {
		v8worker.print(1, ...args)
	},
	warn(...args) {
		v8worker.print(2, ...args)
	},
	error(...args) {
		v8worker.print(3, ...args)
	}
};
`
