package utils

import (

	"os/exec"
	"strings"
)

func wpConf(projName string) bool {
	var base_db, base_fldr string = GenProj(projName)
	var confData Config = GetConf()

	cmd := exec.Command("php/php", "php/wp-cli.phar", "config", "create", "--dbuser="+confData.Db_user, "--dbpass="+confData.Db_pass, "--dbname="+base_db, "--path=htdocs/"+base_fldr)
	_, err := cmd.CombinedOutput()

	Chk_error(err, "Encountering error while executing WP-Cli command ")
	return true
}
func WPInstall(projName string) {
	var _, base_fldr string = GenProj(projName)
	wpConf(projName)
	var confData Config = GetConf()
	cmd := exec.Command("php/php", "php/wp-cli.phar", "core", "install", "--url=http://localhost/"+base_fldr, "--title="+strings.Title(projName), "--admin_user="+confData.Wp_user, "--admin_password="+confData.Wp_pass, "--admin_email="+confData.Wp_email, "--path=htdocs/"+base_fldr)
	_, err := cmd.CombinedOutput()
	Chk_error(err, "Encountering error while executing WP-Cli command Install")
	PrintScrn("'"+projName+"' site created successfully.\nUrl : http://localhost/"+base_fldr+"\nDashboard Url : http://localhost/"+base_fldr+"/wp-admin\nDashboard Username : "+confData.Wp_user+"\nDashboard Password : "+confData.Wp_pass, "green", 0, false)

}
