package v8worker

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestVersion(t *testing.T) {
	println(Version())
}

func TestPrint(t *testing.T) {
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		t.Fatal("shouldn't recieve Message")
		return ""
	})
	err := worker.Execute("code.js", `v8worker.print(0, "ready");`)
	if err != nil {
		t.Fatal(err)
	}
	worker.Dispose()
}

func TestSyntaxError(t *testing.T) {
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		t.Fatal("shouldn't recieve Message")
		return ""
	})

	code := `v8worker.print(0, hello world");`
	err := worker.Execute("codeWithSyntaxError.js", code)
	errorContains(t, err, "codeWithSyntaxError.js")
	errorContains(t, err, "hello")
	worker.Dispose()
}

func TestRequestRecv(t *testing.T) {
	recvCount := 0
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		if len(msg) != 5 {
			t.Fatal("bad msg", msg)
		}
		recvCount++
		return ""
	})

	err := worker.Execute("codeWithRecv.js", `
		v8worker.setRecv(function(mtype, msg) {
			v8worker.print(0, "TestBasic recv byteLength: " + msg + ", " + mtype);
			if (msg.length !== 3) {
				throw Error("bad message");
			}
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
	err = worker.Send(1, "hii")
	if err != nil {
		t.Fatal(err)
	}
	codeWithSend := `
		v8worker.request(2, "12345");
		v8worker.request(2, "12345");
	`
	err = worker.Execute("codeWithSend.js", codeWithSend)
	if err != nil {
		t.Fatal(err)
	}

	if recvCount != 2 {
		t.Fatal("bad recvCount", recvCount)
	}
	worker.Dispose()
}

func TestThrowDuringLoad(t *testing.T) {
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		return ""
	})
	err := worker.Execute("TestThrowDuringLoad.js", `throw Error("bad");`)
	errorContains(t, err, "TestThrowDuringLoad.js")
	errorContains(t, err, "bad")
	worker.Dispose()
}

func TestThrowInRecvCallback(t *testing.T) {
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		return ""
	})
	err := worker.Execute("TestThrowInRecvCallback.js", `
		v8worker.setRecv(function(mtype, msg) {
			v8worker.print(0, "type:" + mtype + ", msg:" + msg)
			throw Error("bad");
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
	err = worker.Send(11, "abcd")
	errorContains(t, err, "TestThrowInRecvCallback.js")
	errorContains(t, err, "bad")
	worker.Dispose()
}

func TestPrintUint8Array(t *testing.T) {
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		return ""
	})
	err := worker.Execute("buffer.js", `
		var uint8 = new Uint8Array(16);
		v8worker.print(0, uint8);
	`)
	if err != nil {
		t.Fatal(err)
	}
	worker.Dispose()
}

func TestMultipleWorkers(t *testing.T) {
	recvCount := 0
	worker1 := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		if len(msg) != 5 {
			t.Fatal("bad message")
		}
		recvCount++
		return ""
	})
	worker2 := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		if len(msg) != 3 {
			t.Fatal("bad message")
		}
		recvCount++
		return ""
	})

	err := worker1.Execute("1.js", `v8worker.request(1, "12345")`)
	if err != nil {
		t.Fatal(err)
	}

	err = worker2.Execute("2.js", `v8worker.request(1, "123")`)
	if err != nil {
		t.Fatal(err)
	}

	if recvCount != 2 {
		t.Fatal("bad recvCount", recvCount)
	}
	worker1.Dispose()
	worker2.Dispose()
}

func TestRequestFromJS(t *testing.T) {
	var captured string
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		captured = msg
		return ""
	})
	code := ` v8worker.request(1, "1234"); `
	err := worker.Execute("code.js", code)
	if err != nil {
		t.Fatal(err)
	}
	if len(captured) != 4 {
		t.Fail()
	}
	worker.Dispose()
}

func TestWorkerBreaking(t *testing.T) {
	worker := New(func(w *Worker, mtype int, msg string, userdata int64) {
		fmt.Printf("%d,%s,%d\n", mtype, msg, userdata)
	}, func(w *Worker, mtype int, msg string) string {
		return ""
	})

	go func(w *Worker) {
		time.Sleep(time.Second)
		w.TerminateExecution()
	}(worker)

	beginTime := time.Now()
	worker.Execute("forever.js", ` while (true) { ; } `)
	fmt.Printf("exec elapse: %v\n", time.Since(beginTime))
	worker.Dispose()
}

func errorContains(t *testing.T, err error, substr string) {
	fmt.Println(err)
	if err == nil {
		t.Fatal("Expected to get error")
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("Expected error to have '%s' in it.", substr)
	}
}
