package server

import (
	"errors"
	"fmt"
	"github.com/lizc2003/gossr/common/tlog"
	"github.com/lizc2003/gossr/common/util"
	v8 "github.com/lizc2003/gossr/v8"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
)

const (
	DIST_DIR_WWW    = "dist_www"
	DIST_DIR_SERVER = "dist_server"
)

type Config struct {
	Host            string      `toml:"host"`
	Log             tlog.Config `toml:"Log"`
	Env             string      `toml:"env"`
	V8MaxCount      int32       `toml:"v8_maxcount"`
	V8LifeTime      int         `toml:"v8_lifetime"`
	JsProjectPath   string      `toml:"js_project_path"`
	StaticUrlPath   string      `toml:"static_url_path"`
	InternalApiHost string      `toml:"internal_api_host"`
	InternalApiIp   string      `toml:"internal_api_ip"`
	InternalApiPort int32       `toml:"internal_api_port"`
	TemplateName    string      `toml:"template_name"`
	ClientCookie    string      `toml:"client_cookie"`
	RedirectOnerror string      `toml:"redirect_onerror"`
	SsrHeaders      []string    `toml:"ssr_headers"`
}

type Server struct {
	RequstMgr          *RequestMgr
	V8Mgr              *v8.V8Mgr
	HostPort           int
	JsProjectPath      string
	Env                string
	IsDevEnv           bool
	ClientCookie       string
	RedirectOnerror    string
	SsrTemplate        string
	SsrHeaders         []string
	tmplateBaseUrl     string
	tmplateAjaxBaseUrl string
}

var ThisServer *Server

func NewServer(c *Config) error {
	jsProjectPath := getJsProjectPath(c.JsProjectPath)
	if jsProjectPath == "" {
		return errors.New("Error: the path of js project is empty.")
	}
	c.JsProjectPath = jsProjectPath

	tmp := strings.Index(c.Host, ":")
	hostPort := int(util.StringToInt64(c.Host[tmp+1:], 0))

	if c.InternalApiPort == 0 {
		c.InternalApiPort = 80
	}

	ThisServer = &Server{
		RequstMgr:       NewRequestMgr(),
		HostPort:        hostPort,
		JsProjectPath:   c.JsProjectPath,
		SsrTemplate:     c.TemplateName,
		ClientCookie:    c.ClientCookie,
		RedirectOnerror: c.RedirectOnerror,
		Env:             c.Env,
		SsrHeaders:      c.SsrHeaders,
		IsDevEnv:        v8.IsDevEnvironment(c.Env),
	}

	v8mgr, err := newV8Mgr(c)
	if err != nil {
		tlog.Fatal(err)
		return err
	}
	ThisServer.V8Mgr = v8mgr

	handler := getHttpHandler(c)
	fmt.Println(util.FormatFullTime(time.Now()), "running ...")
	gracehttp.Serve(&http.Server{Addr: c.Host, Handler: handler})
	return nil
}

func getHttpHandler(c *Config) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()

	localStaticPath := ThisServer.JsProjectPath + DIST_DIR_WWW
	e.Use(GetStaticAndProxyHandler(c.StaticUrlPath, localStaticPath))
	e.StaticFile("/favicon.ico", localStaticPath+"/favicon.ico")
	e.LoadHTMLGlob(c.JsProjectPath + DIST_DIR_SERVER + "/template/*")

	e.NoRoute(HandleSsrRequest)
	return e
}

func getJsProjectPath(jsProjectPath string) string {
	if jsProjectPath == "" {
		return ""
	}
	if jsProjectPath[0] != '/' {
		basepath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return ""
		}
		jsProjectPath = basepath + "/" + jsProjectPath
	}
	if jsProjectPath[len(jsProjectPath)-1] != '/' {
		jsProjectPath += "/"
	}
	return jsProjectPath
}

func newV8Mgr(c *Config) (*v8.V8Mgr, error) {
	serverPath := c.JsProjectPath + DIST_DIR_SERVER + "/"
	serverPathMain := serverPath + "g/"
	vuePath := c.JsProjectPath + "node_modules/"

	v8Conf := v8.V8MgrConfig{
		JsPaths:      []string{serverPathMain, serverPath, vuePath},
		SendCallback: v8SendCallback,
		Env:          c.Env, MaxWorkerCount: c.V8MaxCount, WorkerLifeTime: c.V8LifeTime,
		InternalApiHost: c.InternalApiHost, InternalApiIp: c.InternalApiIp, InternalApiPort: c.InternalApiPort}
	return v8.NewV8Mgr(&v8Conf)
}
