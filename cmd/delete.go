package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"xampress/utils"
	"regexp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteWP)
}

var deleteWP = &cobra.Command{
	Use:   "delete [projectname]",
	Short: "Delete A Exsiting WordPress Site",
	Long:  `Delete Exsiting WordPress Site From XAMPP.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			utils.Start_xampp()
			defer utils.Stop_xampp()
			deleteAction(args[0])
		} else {
			fmt.Println("Please Enter A Project Name\nExample:\nxampress Delete <ProjectName>")
		}
	},
}
func deleteAction(ProjName string){
	var _,base_name string = utils.GenProj(ProjName)
	fmt.Print(base_name)
	checkPath := filepath.Join("htdocs", base_name, "wp-config.php")
	_, err := os.Stat(checkPath)
	var confData,_ = os.ReadFile(checkPath)
	dbNameRegex := regexp.MustCompile(`define\(\s*'DB_NAME'\s*,\s*'([^']+)'\s*\);`)
    dbname := dbNameRegex.FindSubmatch(confData)
	if utils.ProjExist(base_name) && dbname !=nil && err == nil &&utils.ChkDB(string(dbname[1])) {
		os.RemoveAll("htdocs/"+base_name)
		utils.PrintScrn("'"+base_name+"' deleted successfully", "green", 0,false)
		utils.DeleteDBase(string(dbname[1]))
		utils.PrintScrn("'"+string(dbname[1])+"' database deleted successfully", "green",0,false)
		utils.PrintScrn(ProjName+" deleted successfully", "cyan", 0,true)
	}else{
		utils.PrintScrn(ProjName+" name site not found..", "red", 0,false)
	}
}

