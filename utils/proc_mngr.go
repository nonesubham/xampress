package utils

import (
	"os/exec"
	"time"
)
func Stop_xampp(){
	var strt *exec.Cmd= exec.Command("xampp_stop")
	err := strt.Start()
	Chk_error(err, "Unable to stop xampp..")
	
}
func Start_xampp(){
	var strt *exec.Cmd= exec.Command("xampp_start")
	err := strt.Start()
	Chk_error(err, "Server already running")
	time.Sleep(5 *time.Second)
}
