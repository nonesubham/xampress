package cmd

import (
	"fmt"
	"os"
	"xampress/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createWP)
}

var createWP = &cobra.Command{
	Use:   "create [projectname]",
	Short: "Create A New WordPress Site",
	Long:  `Create New WordPress Site On XAMPP.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			utils.Start_xampp()
			defer utils.Stop_xampp()
			createAction(args[0])
		} else {
			fmt.Println("Please Enter A Project Name\nExample:\nxampress create <ProjectName>")
		}
	},
}

func createAction(projName string) {
	var db_name, fldr_name string = utils.GenProj(projName)
	if utils.ChkDB(db_name){
		utils.PrintScrn("'"+db_name+"' named database already exists..", "red",0,false)
		os.Exit(0)
	}
	utils.PrintScrn("Creating '"+fldr_name+"' folder & downloading latest WordPress release...\n","cyan", 0, false)
	utils.WPDown(projName)
	utils.PrintScrn("Creating '"+db_name+"' Database for site\n","cyan", 0, false)
	utils.CreateDBase(db_name)
	utils.PrintScrn("Creating new user and installing Database...\n","cyan", 0, false)
	utils.WPInstall(projName)
}
