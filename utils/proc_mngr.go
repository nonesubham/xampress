package utils

import (
	"os/exec"
	"path/filepath"
	"time"
)

func Stop_xampp(){
	var xamppStop string = filepath.Join(GetConf().Xampp, "xampp_stop")
	var strt *exec.Cmd= exec.Command(xamppStop)
	err := strt.Start()
	Chk_error(err, "Unable to stop xampp..")
	
}
func Start_xampp(){
	var xamppStart string = filepath.Join(GetConf().Xampp, "xampp_start")
	var strt *exec.Cmd= exec.Command(xamppStart)
	err := strt.Start()
	Chk_error(err, "Server already running")
	time.Sleep(5 *time.Second)
}
