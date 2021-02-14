package dockerman

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/docker/docker/api/types/container"
)

func ProcRun(req *RunCodeRequire) error {
	var err error
	if req == nil || req.Code == "" || req.CodeHash =="" || req.InputHash == "" {
		logs.Warn("unexpect params: req=%+v", *req)
		return fmt.Errorf("unexpect params")
	}
	buildPath := getBuildTempPath(req)
	runPath := fmt.Sprintf("%s/%s", buildPath, req.InputHash)
	pathStat := checkPathState(runPath)
	if pathStat == -1 {
		logs.Warn("unexpect error when check path: pathStat=%d runPath=%s", pathStat, runPath)
		return fmt.Errorf("unexpect error")
	}
	if pathStat == 1 {
		logs.Info("skip run process")
		// TODO:return output file
	}
	if pathStat == 2 {
		err = saveStrToFile(req.Input, fmt.Sprintf("%s/input", runPath))
		if err != nil {
			logs.Error("create main.go failed: error=%v", err)
			return err
		}
		logs.Info("create input file success")
	}

	containerConf := &container.Config{
		Image: "alpine:latest",
		WorkingDir: "/workplace",
		NetworkDisabled: true,
		Cmd:   []string{"sh", "-c", "cat input | main"},
	}
	workplaceBind := fmt.Sprintf("%s:/workplace", runPath)
	mainBind := fmt.Sprintf("%s/main:/bin/main", buildPath)
	hostConf := &container.HostConfig{
		Binds:       []string{workplaceBind, mainBind},
	}
	otherConfig := &myConfig{
		AutoRemove:    true,
		MaxTimeSecond: 10,
		ContainerName: "",
	}
	stdOut, errOut, err := runByDocker(containerConf, hostConf, otherConfig)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if stdOut != "" {
		fmt.Println("std-out: ", stdOut)
	}
	if errOut != "" {
		fmt.Println("std-err: ", errOut)
	}

	return nil
}