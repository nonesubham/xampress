/*
Copyright © 2024 Subham nullsubham@gmail.com
*/
package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// clear screen
func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ErrHandler(err error, show_err bool, msg string, raise_issue bool) {
	if err != nil{
		if show_err{
			fmt.Println(err.Error())
		}
		if msg != "" {
			fmt.Println(msg)
		}
		if raise_issue {
			fmt.Println("if you think it's a bug then raise this issue on https://github.com/nonesubham/xampress/issues/new/choose")
		}
		os.Exit(1)
	}

}

// check File exists or not
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

