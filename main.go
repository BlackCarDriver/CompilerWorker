package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"./config"
	"./rpcClient"

	"github.com/astaxie/beego/logs"
	svc "github.com/judwhite/go-svc"
)

//----------------- SVC ---------------
type svcProgram struct{}

func (p *svcProgram) Init(env svc.Environment) error {
	initMain()
	return nil
}

func (p *svcProgram) Start() error {
	go rpcClient.ServerThrift()
	go RegisterRPC()
	return nil
}

func (p *svcProgram) Stop() error {
	return UnRegisterRPC()
}

// ----------------------------------------

func main() {
	prg := &svcProgram{}
	if err := svc.Run(prg); err != nil {
		logs.Error("Run() return error: ", err)
	}
}

func initMain() {
	logs.Info("start initMain...")
	logs.SetLogFuncCall(true)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	if config.ServerConfig.IsTest {
		logs.SetLogger("console")
	} else {
		logs.SetLogger("file", fmt.Sprintf(`{"filename":"%sserver.log", "daily": false}`, config.ServerConfig.LogPath))
	}
}

// 通过http请求注册服务
func RegisterRPC() error {
	logs.Info("start RegisterRPC...")
	var err error
	form := make(url.Values)
	form.Set("ope", "register")
	form.Set("name", config.ServerConfig.S2SName)
	form.Set("url", config.ServerConfig.RequestURI)
	form.Set("tag", config.ServerConfig.Tag)
	form.Set("s2sKey", config.ServerConfig.S2SKey)
	for i := 0; i < 10; i += 1 {
		time.Sleep(time.Duration(i*3) * time.Second)
		var resp *http.Response
		resp, err = http.DefaultClient.PostForm(config.ServerConfig.MasterURI, form)
		if err != nil {
			logs.Warn("post error: time=%d error=%v", i, err)
			continue
		}
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == 207 { // 约定以297作为注册成功的标志
			logs.Info("register result: response=%s", string(bodyBytes))
			return nil
		}
		if resp.StatusCode == 200 {
			logs.Info("register failed, response body=%s", string(bodyBytes))
			continue
		}
		logs.Warn("unexpect response code: time=%d statusCode=%d", i+1, resp.StatusCode)
	}
	logs.Error("give up after try 10 times and failed")
	return err
}

// 通过http请求注销服务
func UnRegisterRPC() error {
	logs.Info("start UnRegisterRPC...")
	var err error
	form := make(url.Values)
	form.Set("ope", "unregister")
	form.Set("name", config.ServerConfig.S2SName)
	form.Set("url", config.ServerConfig.RequestURI)
	form.Set("tag", config.ServerConfig.Tag)
	form.Set("s2sKey", config.ServerConfig.S2SKey)
	for i := 0; i < 5; i += 1 {
		var resp *http.Response
		resp, err = http.DefaultClient.PostForm(config.ServerConfig.MasterURI, form)
		if err != nil {
			logs.Warn("post error: time=%d error=%v", i, err)
			continue
		}
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == 207 {
			logs.Info("unregister result: response=%s", string(bodyBytes))
			return nil
		}
		if resp.StatusCode == 200 {
			logs.Info("register failed, response body=%s", string(bodyBytes))
			continue
		}
		logs.Warn("unexpect response code: time=%d statusCode=%d", i+1, resp.StatusCode)
	}
	logs.Error("give up after try 10 times and failed")
	return err
}
