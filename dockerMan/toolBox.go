package dockerman

import (
	"bufio"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"os"
)

// return 1:alreadyExist 2:notExistAndCreated -1:notAFloderOrOtherError
func checkPathState(path string) int {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return 1
	}
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			logs.Error("make direct failed: error=%v path=%s", err, path)
			return -1
		}
		return 2
	}
	return -1
}

// write src into a file, if the old file is exist, will just skip
func saveStrToFile(str, path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file '%s' is alread exist", path)
	} else if os.IsNotExist(err) == false {
		return err
	}
	err := ioutil.WriteFile(path, []byte(str), 0644)
	return err
}

//read a file and parse it into a string 📝
func ParseFile(path string) (text string, err error) {
	file, err := os.Open(path)
	if err != nil {
		logs.Warn("Open %s fall: error=%v", path, err)
		return "", err
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		logs.Warn("ioutil.ReadAll fall : %v", err)
		return "", err
	}
	return string(bytes), nil
}
