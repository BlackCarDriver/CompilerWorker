package rpcClient

import (
	"baseService"
	"context"

	"github.com/astaxie/beego/logs"
)

type myCodeRunner struct{}

func (r *myCodeRunner) Ping(ctx context.Context, str string) (string, error) {
	logs.Info("ping receive: %s", str)
	return str, nil
}

func (r *myCodeRunner) BuildGo(ctx context.Context) (*baseService.CommomResp, error) {
	logs.Info("buildGo...")
	resp := baseService.CommomResp{
		Status:  0,
		Msg:     "OK",
		Payload: nil,
	}
	return &resp, nil
}

func (r *myCodeRunner) BuildCpp(ctx context.Context) (*baseService.CommomResp, error) {
	logs.Info("buildCpp...")
	resp := baseService.CommomResp{
		Status:  0,
		Msg:     "OK",
		Payload: nil,
	}
	return &resp, nil
}

func (r *myCodeRunner) Run(ctx context.Context) (*baseService.CommomResp, error) {
	logs.Info("run...")
	resp := baseService.CommomResp{
		Status:  0,
		Msg:     "OK",
		Payload: nil,
	}
	return &resp, nil
}
