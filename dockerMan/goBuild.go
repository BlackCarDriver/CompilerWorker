package dockerman

import (
	"../config"
	"errors"
	
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/docker/docker/api/types/container"
)

// GoBuild build go code and run the process
func GoBuild(req *RunCodeRequire) (stdErr, stdOut string, err error) {
	if req == nil || req.Code == "" || req.CodeHash == "" || req.InputHash == "" {
		logs.Warn("unexpect params: req=%+v", *req)
		err = errors.New("unexpect params")
		return
	}
	tempPath := getBuildTempPath(req)
	pathState := checkPathState(tempPath)
	logs.Debug("tempPath=%s pathStat=%d", tempPath, pathState)

	if pathState == -1 {
		logs.Error("something go worng...") // check or create floder failed
		err = errors.New("something go worng, please check")
	}
	if pathState == 1 { // already create before
		logs.Info("skip build go code")
		stdErr, err = ParseFile(fmt.Sprintf("%s/stderr", tempPath))
		if err!=nil || stdErr != "" {
			return
		}
		stdOut, err = ParseFile(fmt.Sprintf("%s/stdout", tempPath))
		return
	}
	if pathState == 2 { // new floder for new code
		err = saveStrToFile(req.Code, fmt.Sprintf("%s/main.go", tempPath))
		if err != nil {
			logs.Error("create main.go failed: error=%v", err)
			return
		}
		logs.Info("create main.go success")
	}

	// build main.go
	containerConf := &container.Config{
		Image:           "golang:alpine",
		Env:             []string{"GOARCH=amd64"},
		WorkingDir:      "/workplace",
		NetworkDisabled: true,
		Cmd:             []string{"sh", "-c", "go build main.go 1>stdout 2>stderr"},
	}
	bindGoPath := fmt.Sprintf("%s:/go", config.ServerConfig.GoPath)
	workplaceBin := fmt.Sprintf("%s:/workplace", tempPath)
	hostConf := &container.HostConfig{
		Binds: []string{bindGoPath, workplaceBin},
	}
	otherConfig := &myConfig{
		AutoRemove:    true,
		MaxTimeSecond: 10,
		ContainerName: "",
	}
	_, _, err = runByDocker(containerConf, hostConf, otherConfig)
	if err != nil {
		logs.Error(err)
		return
	}
	stdOut, err = ParseFile(fmt.Sprintf("%s/stdout", tempPath))
	if err!=nil{
		logs.Warn("read stdout failed: error=%v", err)
		return
	}
	stdErr, err = ParseFile(fmt.Sprintf("%s/stderr", tempPath))
	if err!=nil{
		logs.Warn("read stderr failed: error=%v", err)
		return
	}
	logs.Debug("stdOut=%q  stdError=%q", stdOut, stdErr)
	return
}
