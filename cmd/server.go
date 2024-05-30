package cmd

import (
	"fmt"
	"xampress/utils"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(server)
}

var server = &cobra.Command{
	Use:   "server [start/stop]",
	Short: "Control XAMPP server",
	Long:  `Control MYSQL & Apache server of XAMPP.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			serverAction(args[0])
		} else {
			fmt.Println("Please enter a valid command\nExample:\nxampress server start/stop")
		}
	},
}
func serverAction(usr_inp string){
	if usr_inp == "start"{
		utils.Start_xampp()
		utils.PrintScrn("XAMPP started successfully..","green", 0, false)
	} else if usr_inp =="stop"{
		utils.Stop_xampp()
		utils.PrintScrn("XAMPP stopped successfully..","green", 0, false)
	}else{
		fmt.Println("Please enter a valid command\nExample:\nxampress server start/stop")
	}
}