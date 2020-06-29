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

package main

import (
	"fmt"
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/common/util"
	logic "github.com/lizc2003/gossr/server"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	var c logic.Config
	if !util.ParseConfig("./conf/gossr-dev.toml", &c) {
		return
	}
	tlog.Init(c.Log)

	//go func() {
	//	tlog.Info(http.ListenAndServe("0.0.0.0:32123", nil))
	//}()

	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())

	if err := logic.NewServer(&c); err == nil {
	} else {
		fmt.Println(err)
	}

	tlog.Close()
}
