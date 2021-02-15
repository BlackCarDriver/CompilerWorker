package rpcClient

import (
	dockerman "../dockerMan"
	"baseService"
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
)

const hashLen = 10

type respPayload struct {
	StdErr string `json:"stdErr"`
	StdOut string `json:"stdOut"`
}

type myCodeRunner struct{}

func (r *myCodeRunner) Ping(ctx context.Context, str string) (string, error) {
	logs.Info("ping receive: %s", str)
	return str, nil
}

func (r *myCodeRunner) BuildGo(ctx context.Context, code string, input string) (resp *baseService.CommomResp, err error) {
	logs.Info("buildGo...")
	var payload respPayload
	resp = &baseService.CommomResp{
		Status: -1,
		Msg:    "OK",
	}
	req := dockerman.RunCodeRequire{
		Type:      "GO",
		Code:      code,
		Input:     input,
		CodeHash:  dockerman.GetMD5KeyN(code, 10),
		InputHash: dockerman.GetMD5KeyN(input, hashLen),
	}
	for loop := true; loop; loop = false {
		payload.StdErr, payload.StdOut, err = dockerman.GoBuild(&req)
		logs.Info("GoBuild: error=%v stdErr=%q stdOut=%q", err, payload.StdErr, payload.StdOut)
		if err != nil {
			break
		}
		payload.StdErr, payload.StdOut, err = dockerman.ProcRun(&req)
		logs.Info("ProcRun result: error=%v stdErr=%q stdOut=%q", err, payload.StdErr, payload.StdOut)
		if err != nil {
			break
		}
	}
	if err != nil {
		resp.Status = -1
		resp.Msg = fmt.Sprint(err)
	}
	resp.Payload, _ = json.Marshal(payload)

	return
}

func (r *myCodeRunner) BuildCpp(ctx context.Context, code string, input string) (*baseService.CommomResp, error) {
	logs.Info("buildCpp...")
	resp := baseService.CommomResp{
		Status:  0,
		Msg:     "OK",
		Payload: nil,
	}
	logs.Info("code=%s  input=%s", code, input)
	return &resp, nil
}

func (r *myCodeRunner) Run(ctx context.Context, codeType, hash, input string) (resp *baseService.CommomResp, err error) {
	logs.Info("run...")
	var payload respPayload

	runReq := &dockerman.RunCodeRequire{
		Type:      codeType,
		CodeHash:  hash,
		InputHash: dockerman.GetMD5KeyN(input, hashLen),
	}
	resp = &baseService.CommomResp{
		Status:  0,
		Msg:     "OK",
		Payload: nil,
	}
	payload.StdErr, payload.StdOut, err = dockerman.ProcRun(runReq)

	if err != nil {
		resp.Status = -1
		resp.Msg = fmt.Sprint(err)
	}
	resp.Payload, _ = json.Marshal(payload)
	return resp, nil
}
