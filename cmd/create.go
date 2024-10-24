/*
Copyright © 2024 Subham nullsubham@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"path"

	// "path"
	"xampress/helpers"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) !=0 {
			crtFn(args)
		}else{
			fmt.Println("Please provide site name in order to create site!")
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
func crtFn(projNames []string) {
	folderName, dbName, siteName := helpers.CleanName(projNames)
	xamppConfig, err := helpers.GetConf("xampp")
	helpers.ErrHandler(err, true, "", true)

	xamppSettings, ok := xamppConfig.(map[string]interface{})
	if !ok {
		fmt.Printf("Invalid value: %v\nUpdate it with 'config' command.\n", xamppConfig)
		return
	}

	xamppPath, ok := xamppSettings["xampp"].(string)
	if !ok {
		fmt.Println("Invalid value for 'xampp' in configuration. It must be a string.")
		return
	}

	htdocsPath, ok := xamppSettings["htdocs"].(string)
	if !ok {
		fmt.Println("Invalid value for 'htdocs' in configuration. It must be a string.")
		return
	}

	projectDir := path.Join(xamppPath, htdocsPath, folderName)
	if helpers.DirExists(projectDir) {
		fmt.Println(folderName, "named folder already exists, please use another name.")
		return
	}

	fmt.Println("Creating directory at", projectDir, dbName, siteName)
	helpers.ErrHandler(os.Mkdir(projectDir, os.ModePerm), true, "", false)
}


