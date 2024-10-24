/*
Copyright © 2024 Subham nullsubham@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"xampress/helpers"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "xampress",
	Short: "Xampress, a tool that automates WordPress site management in XAMPP",
	Long: `Xampress a tool that automates WordPress site management in XAMPP,
enabling the creation of new sites and the installation of WordPress with a single click.
Ideal for achieving efficiency and ease in workflow.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
}

func init() {
	if isExist, err := helpers.ChkXamppDirs(); !isExist {
		if err != nil {
		var pcmd string = ""
			if err != "binaries" {
				pcmd = "dir"
			}
			fmt.Println("XAMPP or some important XAMPP's directories not found at defaut installation location. Please update dir",err,"at config.yaml with config command\neg :", "xampress config set --"+pcmd+"-"+err.(string), "<custom dir location of", err,">")
		}else{
			fmt.Println("Encountering Error while checking XAMPP and XAMPP's default dirs.")
		}
		os.Exit(1)
	}
}
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}


