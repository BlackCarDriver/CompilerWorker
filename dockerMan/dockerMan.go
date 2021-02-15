package dockerman

import (
	"../config"
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var (
	cli *client.Client
)

func InitDockerClient() (err error) {
	cli, err = client.NewClientWithOpts(client.WithVersion("1.39"))
	if err != nil {
		logs.Emergency("Init client fall: error=%v", err)
		os.Exit(1)
	}
	return nil
}

// TestGoBuild test if docker fucntion is normal on the machine when the program init
func TestDockerFunction() error {
	// dockerman.TestBuild()
	var exampleCode = `package main
	import "fmt"
	func main(){
		fmt.Println("SUCCESS")
	}`
	req := RunCodeRequire{
		Type:      "GO",
		Code:      exampleCode,
		Input:     "xianjinrong 22",
		CodeHash:  GetMD5KeyN(exampleCode, 10),
		InputHash: GetMD5KeyN("xianjinrong 22", 10),
	}
	_, _, err := GoBuild(&req)
	if err != nil {
		logs.Error("Gobuild not pass: error=%v", err)
		return err
	}
	_, _, err = ProcRun(&req)
	if err != nil {
		logs.Error("Gobuild not pass: error=%v", err)
		return err
	}
	logs.Info("test docker function pass")
	return nil
}

// ---------------- some of the struct --------------------------------

// RunCodeRequire repercent the require of build go code
type RunCodeRequire struct {
	Type      string `json:"type"` //[GO|CPP]
	Code      string `json:"code"`
	Input     string `json:"input"`
	CodeHash  string `json:""`
	InputHash string `json:"inputHash"`
}

type myConfig struct {
	MaxTimeSecond int
	ContainerName string
	AutoRemove    bool
}

// ---------------- base docker tool function -------------------------

//create a docker container and return the container id
func createContainer(config *container.Config, hostConfig *container.HostConfig) (containerID string, err error) {
	if cli == nil {
		return "", errors.New("docker client is not init")
	}
	resp, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

//start the docker container
func startContainer(containerID string) error {
	if cli == nil {
		return errors.New("docker client is not init")
	}
	return cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
}

//kill and remove a docker ocntainer
func removeContainer(containerID string) error {
	if cli == nil {
		return errors.New("docker client is not init")
	}
	return cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
}

//get the std-out and std-error of a container
func getOuput(containerID string) (stdout, stderr string, err error) {
	if cli == nil {
		return "", "", errors.New("docker client is not init")
	}
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", "", fmt.Errorf("Create client fall: %v", err)
	}
	option := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}
	stdoutRC, err := cli.ContainerLogs(ctx, containerID, option)
	if err != nil {
		return "", "", fmt.Errorf("Error happen when geting container's logs of: %v", err)
	}
	defer stdoutRC.Close()
	stdbuf := bytes.NewBufferString("")
	errbuf := bytes.NewBufferString("")
	_, err = stdcopy.StdCopy(stdbuf, errbuf, stdoutRC)
	if err != nil {
		return "", "", fmt.Errorf("StdCopy fall: %v", err)
	}
	return stdbuf.String(), errbuf.String(), nil
}

// ---------------------- docker snak function ----------------------

func removeContainer2(containerID string) {
	err := removeContainer(containerID)
	if err != nil {
		logs.Error("remove container error: ID=%s error=%v", containerID, err)
	} else {
		logs.Info("remove container success: ID=%s", containerID)
	}
}

//run a docker by given config and return the std-ouput and std-error
func runByDocker(config *container.Config, hostConfig *container.HostConfig, otherConfig *myConfig) (stdOut, errOut string, err error) {
	ctx := context.Background()
	var cli *client.Client
	for loop := true; loop; loop = false {
		// create client
		timestamp := time.Now()
		cli, err = client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			logs.Error("create client failed: error=%v", err)
			break
		}
		logs.Info("create client success, use %d ms", time.Since(timestamp).Milliseconds())
		defer cli.Close()
		cli.NegotiateAPIVersion(ctx)

		//create container
		timestamp = time.Now()
		var resp container.ContainerCreateCreatedBody
		resp, err = cli.ContainerCreate(ctx, config, hostConfig, nil, otherConfig.ContainerName)
		if err != nil {
			logs.Error("create container failed: error=%v", err)
			break
		}
		logs.Info("create container success, use %d ms", time.Since(timestamp).Milliseconds())
		if otherConfig.AutoRemove {
			defer removeContainer2(resp.ID)
		}

		//run container
		timestamp = time.Now()
		err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
		if err != nil {
			logs.Error("run container failed: error=%v", err)
			break
		}
		logs.Info("start client success, use %d ms", time.Since(timestamp).Milliseconds())
		timestamp = time.Now()

		//wait the container stop
		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err = <-errCh: //fail to run the container
			logs.Error("run container failed: error=%v", err)
			break

		case <-time.After(time.Duration(otherConfig.MaxTimeSecond) * time.Second): //container run out of time
			logs.Error("run container time expired: maxTime=%d", otherConfig.MaxTimeSecond)
			err = fmt.Errorf("run time expired: maxTime=%d", otherConfig.MaxTimeSecond)
			break

		case <-statusCh: //scuess run and stop the container
			stdoutRC, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
			if err != nil {
				logs.Error("get logs failed: error=%v", err)
				break
			}
			defer stdoutRC.Close()
			stdbuf := bytes.NewBufferString("")
			errbuf := bytes.NewBufferString("")
			_, err = stdcopy.StdCopy(stdbuf, errbuf, stdoutRC)
			if err != nil {
				logs.Error("read ouput failed: error=%v", err)
				break
			}
			stdOut = stdbuf.String()
			errOut = errbuf.String()
		}
		logs.Info("run container stop, use %d ms", time.Since(timestamp).Milliseconds())
	}
	if err != nil {
		logs.Error("run container fail: ContainerConfig=%+v hostConfig=%+v otherConfig=%+v error=%v", *config, *hostConfig, otherConfig, err)
		return "", "", err
	}
	return stdOut, errOut, nil
}

// ---------------------- other snak function --------------------

// return the path where the program should build to
func getBuildTempPath(req *RunCodeRequire) (path string) {
	md5Key := req.CodeHash
	if md5Key == "" {
		logs.Warn("unexpect params: req=%+v", *req)
		path = "default"
	}
	switch req.Type {
	case "GO":
		path = fmt.Sprintf("%sGo/%s", config.ServerConfig.BuildResultPath, md5Key)
	case "CPP":
		path = fmt.Sprintf("%sCPP/%s", config.ServerConfig.BuildResultPath, md5Key)
	default:
		logs.Error("unexpect params: req=%+v", *req)
		path = fmt.Sprintf("%sUnknow/%s", config.ServerConfig.BuildResultPath, md5Key)
	}
	return path
}

// ---------------------- demo -------------------------------

// TestRun1 test usage
func TestRun1() {
	containerConf := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"/bin/sh", "-c", "ls /"},
	}
	hostConf := &container.HostConfig{
		OomScoreAdj: 1000,
		Resources:   container.Resources{Memory: 104857600, NanoCPUs: 50000000},
	}
	otherConfig := &myConfig{
		AutoRemove:    true,
		MaxTimeSecond: 10,
		ContainerName: "testalpine",
	}
	stdOut, errOut, err := runByDocker(containerConf, hostConf, otherConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	if stdOut != "" {
		fmt.Println("std-out: ", stdOut)
	}
	if errOut != "" {
		fmt.Println("std-err: ", errOut)
	}
}

// TestBuild test build go code
func TestBuild() {
	containerConf := &container.Config{
		Image:           "golang:alpine",
		Env:             []string{"CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64"},
		WorkingDir:      "/workplace",
		NetworkDisabled: true,
		Cmd:             []string{"go", "build", "main.go"},
	}
	hostConf := &container.HostConfig{
		Binds: []string{"/home/driver/GoPath:/go", "/home/driver/tempWorkplace:/workplace"},
	}
	otherConfig := &myConfig{
		AutoRemove:    true,
		MaxTimeSecond: 10,
		ContainerName: "",
	}
	stdOut, errOut, err := runByDocker(containerConf, hostConf, otherConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	if stdOut != "" {
		fmt.Println("std-out: ", stdOut)
	}
	if errOut != "" {
		fmt.Println("std-err: ", errOut)
	}
}

// GetMD5KeyN create a hash code
func GetMD5KeyN(anyMsg interface{}, n int) string {
	md5Encoder := md5.New()
	md5Encoder.Write([]byte(fmt.Sprint(anyMsg)))
	md5Value := fmt.Sprintf("%x", md5Encoder.Sum(nil))
	if n <= 0 || n >= len(md5Value) {
		logs.Warning("unexpect length: length=%d\n", n)
		return md5Value
	}
	return md5Value[0:n]
}
