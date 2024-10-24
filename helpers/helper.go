/*
Copyright © 2024 Subham nullsubham@gmail.com
*/
package helpers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
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

// check File/Folder exists or not
func DirExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// check internet is connected or not
func CheckConnection() bool {
	resp, err := http.Get("https://httpbin.org/get")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
// return folder name and database name, site name from proj_name
func CleanName(proj_name []string) (string, string, string) {
	var fldr_name string = strings.Join(proj_name, "-")
	var db_name string = strings.ReplaceAll(fldr_name, "-","_")
	return fldr_name, db_name, strings.Join(proj_name, " ")
}

//checks for xampp importants dirs exist at location according to config.yaml
func ChkXamppDirs() (bool, any) {

	bins := []string{"php", "apache", "mysql"}

	xamppDirs, err := GetConf("xampp")
	ErrHandler(err, true, "", false)

	allDirs, ok := xamppDirs.(map[string]interface{})
	if !ok {
		return false, nil
	}

	baseDir, ok := allDirs["xampp"].(string)
	if !ok {
		return false, nil
	}

	for _, binDir := range bins {
		var fBinDir string = path.Join(baseDir, binDir)
		if !DirExists(fBinDir) {
			return false, "binaries"
		}
	}

	// Check for other directories
	for name, dir := range allDirs {
		var fDir string = path.Join(baseDir, dir.(string))
		if name != "binaries" && !DirExists(fDir) {
			return false, path.Base(fDir)
		}
	}

	return true, nil
}
