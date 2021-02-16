package dockerman

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/docker/docker/api/types/container"
)

func ProcRun(req *RunCodeRequire) (stdErr, stdOut string, err error) {
	if req == nil || req.Type == "" || req.CodeHash == "" || req.InputHash == "" {
		logs.Warn("unexpect params: req=%+v", *req)
		err = fmt.Errorf("unexpect params")
		return
	}
	buildPath := getBuildTempPath(req)
	runPath := fmt.Sprintf("%s/%s", buildPath, req.InputHash)
	logs.Debug("RunCodeRequire=%+v", *req)
	logs.Debug("buildPath=%s runPath=%s", buildPath, runPath)

	pathStat := checkPathState(runPath)
	if pathStat == -1 {
		logs.Warn("unexpect error when check path: pathStat=%d runPath=%s", pathStat, runPath)
		err = errors.New("unexpect error")
		return
	}
	if pathStat == 1 {
		logs.Info("skip run process")
		stdErr, err = ParseFile(fmt.Sprintf("%s/stderr", runPath))
		if err!=nil || stdErr != "" {
			return
		}
		stdOut, err = ParseFile(fmt.Sprintf("%s/stdout", runPath))
		return
	}
	if pathStat == 2 {
		err = saveStrToFile(req.Input, fmt.Sprintf("%s/input", runPath))
		if err != nil {
			logs.Error("create main.go failed: error=%v", err)
			return
		}
		logs.Info("create input file success")
	}

	containerConf := &container.Config{
		Image:           "ubuntu:latest",
		WorkingDir:      "/workplace",
		NetworkDisabled: true,
		Cmd:             []string{"sh", "-c", "cat input | main 1>stdout 2>stderr"},
	}
	workplaceBind := fmt.Sprintf("%s:/workplace", runPath)
	mainBind := fmt.Sprintf("%s/main:/bin/main", buildPath)
	hostConf := &container.HostConfig{
		Binds: []string{workplaceBind, mainBind},
	}
	otherConfig := &myConfig{
		AutoRemove:    true,
		MaxTimeSecond: 10,
		ContainerName: "",
	}
	_, _, err = runByDocker(containerConf, hostConf, otherConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	stdErr, _ = ParseFile(fmt.Sprintf("%s/stderr", runPath))
	stdOut, _ = ParseFile(fmt.Sprintf("%s/stdout", runPath))
	logs.Debug("stdErr=%q  stdOut=%q", stdErr, stdOut)

	return
}
