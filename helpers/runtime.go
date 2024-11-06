package helpers

import (
	"fmt"
	"os/exec"
	"runtime"

	"rxcsoft.cn/utils/logger"
)

var log = logger.New()

// CurrentOSVer get current runtime OS
func CurrentOSVer() string {
	return runtime.GOOS
}

// LookPath get bin path to where is available
func LookPath(binName string) (string, error) {
	path, err := exec.LookPath(binName)
	if err != nil {
		log.Fatal(fmt.Sprintf("not found %v", binName))
		return path, err

	}

	return path, nil
}

// OSShellCommand get and set shell command for win/linux/osx
func OSShellCommand() string {
	var cmd string
	switch CurrentOSVer() {
	case "windows":
		path, err := LookPath("sh")
		if err != nil {
			log.Fatal(err)
		}
		cmd = path
	default:
		cmd = "/bin/bash"
	}

	return cmd
}
