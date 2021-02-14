package dockerman

import (
	"errors"
	"fmt"
	"../config"
	"github.com/docker/docker/api/types/container"
	"github.com/astaxie/beego/logs"
)

// GoBuild build go code
func GoBuild(req *RunCodeRequire) error {
	if req == nil || req.Code == "" || req.CodeHash =="" || req.InputHash == "" {
		logs.Warn("unexpect params: req=%+v", *req)
		return fmt.Errorf("unexpect params")
	}
	tempPath := getBuildTempPath(req)
	pathState := checkPathState(tempPath)
	if pathState == -1 {
		logs.Error("something go worng...") // check or create floder failed
		return errors.New("something go worng, please check")
	}
	if pathState == 1 { // already create before
		logs.Info("skip build go code")
		return nil
	}
	if pathState == 2 { // new floder for new code
		err := saveStrToFile(req.Code, fmt.Sprintf("%s/main.go", tempPath))
		if err != nil {
			logs.Error("create main.go failed: error=%v", err)
			return err
		}
		logs.Info("create main.go success")
	}

	// build main.go
	containerConf := &container.Config{
		Image: "golang:alpine",
		Env: []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64"},
		WorkingDir: "/workplace",
		NetworkDisabled: true,
		Cmd:   []string{"go", "build", "main.go"},
	}
	bindGoPath := fmt.Sprintf("%s:/go", config.ServerConfig.GoPath)
	workplaceBin := fmt.Sprintf("%s:/workplace", tempPath)
	hostConf := &container.HostConfig{
		Binds:       []string{bindGoPath, workplaceBin},
	}
	otherConfig := &myConfig{
		AutoRemove:    true,
		MaxTimeSecond: 10,
		ContainerName: "",
	}
	stdOut, errOut, err := runByDocker(containerConf, hostConf, otherConfig)
	if err != nil {
		logs.Info(err)
		return err
	}
	if stdOut != "" {
		logs.Info("std-out: ", stdOut)
	}
	if errOut != "" {
		logs.Info("std-err: ", errOut)
	}

	return nil
}
