package cmd

import (
	"os"
	"strconv"
	"strings"
	"xampress/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listWP)
}

var listWP = &cobra.Command{
	Use:   "list",
	Short: "Get all installed WordPress site lists..",
	Long:  `Get all installed WordPress site lists from 'htdocs'`,
	Run: func(cmd *cobra.Command, args []string) {
		listAction()

	},
}

func listAction() {
	folders, ht_err := os.ReadDir("htdocs")
	utils.Chk_error(ht_err, "Can't able to fetch folders from 'htdocs'")
	var sites []string
	for _, folder := range folders{
		if _, stat_err := os.Stat("htdocs/"+folder.Name()+"/wp-config.php"); !os.IsNotExist(stat_err){
			sites= append(sites, folder.Name())
		}
	}
	utils.PrintScrn("There are total "+ strconv.Itoa(len(sites))+" WordPress Sites..\n", "yellow", 0, false)
	for i,f := range sites{

		utils.PrintScrn(strconv.Itoa(i+1)+". ", "white", 0, false)
		utils.PrintScrn(strings.Title(strings.ReplaceAll(f,"-"," "))+"\n", "magenta", 0, false)
	}
}