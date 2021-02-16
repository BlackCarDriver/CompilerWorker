package dockerman

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/docker/docker/api/types/container"
)

// CppBuild build C or C++ code
func CppBuild(req *RunCodeRequire) (stdErr, stdOut string, err error) {
	if req == nil || req.Code == "" || req.CodeHash == "" || req.InputHash == "" {
		logs.Warn("unexpect params: req=%+v", *req)
		err = errors.New("unexpect params")
		return
	}
	tempPath := getBuildTempPath(req)
	pathState := checkPathState(tempPath)
	logs.Debug("tempPath=%s pathStat=%d", tempPath, pathState)
	fileName := "main.cpp"
	command := "g++"
	if req.Type == "C" {
		fileName = "main.c"
		command = "gcc"
	}

	if pathState == -1 {
		logs.Error("something go worng...") // check or create floder failed
		err = errors.New("something go worng, please check")
	}
	if pathState == 1 { // already create before
		logs.Info("skip build c/c++ code")
		stdErr, err = ParseFile(fmt.Sprintf("%s/stderr", tempPath))
		if err!=nil || stdErr != "" {
			return
		}
		stdOut, err = ParseFile(fmt.Sprintf("%s/stdout", tempPath))
		return
	}
	if pathState == 2 { // new floder for new code
		err = saveStrToFile(req.Code, fmt.Sprintf("%s/%s", tempPath, fileName))
		if err != nil {
			logs.Error("create main.go failed: error=%v", err)
			return
		}
		logs.Info("create %s success", fileName)
	}

	// build main.c or main.cpp
	containerConf := &container.Config {
		Image:           "gcc:latest",
		WorkingDir:      "/workplace",
		NetworkDisabled: true,
		Cmd:             []string{"sh", "-c", fmt.Sprintf("%s -o main %s 1>stdout 2>stderr", command, fileName)},
	}
	workplaceBin := fmt.Sprintf("%s:/workplace", tempPath)
	hostConf := &container.HostConfig{
		Binds: []string{workplaceBin},
	}
	otherConfig := &myConfig{
		AutoRemove:    true,
		MaxTimeSecond: 10,
		ContainerName: "",
	}
	// run the program
	_, _, err = runByDocker(containerConf, hostConf, otherConfig)
	if err != nil {
		logs.Info(err)
		return
	}
	stdErr, err = ParseFile(fmt.Sprintf("%s/stderr", tempPath))
	if err!=nil {
		logs.Warn("read stderr failed: error=%v", err)
		return
	}
	stdOut, err = ParseFile(fmt.Sprintf("%s/stdout", tempPath))
	if err!=nil{
		logs.Warn("read stdout failed: error=%v", err)
		return
	}

	logs.Debug("stdOut=%q  stdError=%q", stdOut, stdErr)
	return
}
