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
