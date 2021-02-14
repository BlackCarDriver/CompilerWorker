package dockerman

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/astaxie/beego/logs"
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
