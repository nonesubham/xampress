package cmd

import (
	"encoding/json"
	"os"
	"runtime"
	"xampress/utils"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "XamPress [Create/Delete/Clone] <Project Name>",
		Short: "XamPress is a CLI tool for managing WordPress sites on Xampp",
		Long: `Xampress automates WordPress project management within the XAMPP environment, offering a user-friendly command-line interface (CLI) for seamless automation. This tool simplifies the installation process, swiftly configuring WordPress instances on XAMPP with ease. With Xampress, developers can effortlessly create and manage WordPress projects, accelerating development workflows and boosting productivity.
		
		view this project on github: https://github.com/nonesubham/xampress
		`,
	}
)

func init() {
	var xamp_dir string

	switch runtime.GOOS {
	case "windows":
		xamp_dir = "C:/xampp"
	case "darwin":
		xamp_dir = "/Applications/XAMPP/"
	case "linux":
		xamp_dir = "/Applications/XAMPP/"
	}
	os.Chdir(xamp_dir)
	if _, err := os.Stat("xampress/config.json"); os.IsNotExist(err) {
		file, errj := os.Create("xampress/config.json")
		utils.Chk_error(errj, "Error while creating 'config.json'")
		defer file.Close()
		defConf := (utils.Config{Db_user: "root", Db_pass: "", Wp_user: "root", Wp_pass: "root", Wp_email: "example@domain.com", Xampp: xamp_dir})
		enc_data, _ := json.Marshal(defConf)
		_, errw := file.Write(enc_data)
		utils.Chk_error(errw, "Encountering while writing data in 'config.json'")
	}
	confData := utils.GetConf()
	rootCmd.AddCommand(configCMD)
	configCMD.Flags().StringVar(&sql_user, "sql-user", confData.Db_user, "Set MySQL username")
	configCMD.Flags().StringVar(&sql_pass, "sql-pass", confData.Db_pass, "Set MySQL password")
	configCMD.Flags().StringVar(&wp_user, "wp-user", confData.Wp_user, "Set WordPress username")
	configCMD.Flags().StringVar(&wp_pass, "wp-pass", confData.Wp_pass, "Set WordPress password")
	configCMD.Flags().StringVar(&wp_email, "wp-email", confData.Wp_email, "Set WordPress email")
	configCMD.Flags().StringVar(&xampp_path, "xampp-dir", confData.Xampp, "Set XAMPP path")
}

func Execute() error {
	return rootCmd.Execute()
}
