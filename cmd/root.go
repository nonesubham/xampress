package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"xampress/utils"
    "runtime"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "XamPress [Create/Delete/Clone] <Project Name>",
		Short: "XamPress is a CLI tool for managing WordPress sites on Xampp",
		Long: `Xampress automates WordPress project management within the XAMPP environment, offering a user-friendly command-line interface (CLI) for seamless automation. This tool simplifies the installation process, swiftly configuring WordPress instances on XAMPP with ease. With Xampress, developers can effortlessly create and manage WordPress projects, accelerating development workflows and boosting productivity.
		
		view this project on github: https://github.com/nonesubham/XamPress
		`,
	}
)
func init() {
    var xamp_path string
    if runtime.GOOS =="windows"{
        xamp_path = "C:/xampp"
    }else if runtime.GOOS == "linux"{
        xamp_path = "/opt/lampp"
    }else if runtime.GOOS=="darwin"{
        xamp_path = "/Applications/XAMPP"
    }
	os.Chdir(xamp_path)
	if _, err := os.Stat("xampress/config.json"); os.IsNotExist(err){
        file, _ := os.Create("xampress/config.json")
        // utils.Chk_error(errj,)
        fmt.Print(os.Getwd())
        defer file.Close()
        defConf := (utils.Config{Db_user: "root", Db_pass: "", Wp_user: "root", Wp_pass: "root", Wp_email: "example@domain.com"})
        enc_data,_ := json.Marshal(defConf)
        _, errw := file.Write(enc_data)
        utils.Chk_error(errw, "Encountering while writing data in 'config.json'")
    }
    confData := utils.GetConf()
    rootCmd.AddCommand(configCMD)
    configCMD.Flags().StringVar(&sql_user, "sql-user", confData.Db_user, "Set MySQL username")
    configCMD.Flags().StringVar(&sql_pass, "sql-pass", confData.Db_pass, "Set MySQL password")
    configCMD.Flags().StringVar(&wp_user, "wp-user", confData.Wp_user, "Set WordPress username")
    configCMD.Flags().StringVar(&wp_pass, "wp-pass", confData.Wp_pass, "Set WordPress password")
    configCMD.Flags().StringVar(&wp_email, "wp-email",confData.Wp_email, "Set WordPress email")
}

func Execute() error {
	return rootCmd.Execute()
}
