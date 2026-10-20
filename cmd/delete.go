package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"xampress/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteWP)
}

var deleteWP = &cobra.Command{
	Use:   "delete <projectname>",
	Short: "Delete an existing local WordPress or Laravel project",
	Long:  `Deletes an existing WordPress or Laravel project directory and drops its associated MySQL database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			if err := utils.EnsureMySQLRunning(); err != nil {
				utils.Chk_error(err, err.Error())
			}
			deleteAction(args[0])
		} else {
			fmt.Println("Please enter a project name.\nExample:\n  xampress delete <ProjectName>")
		}
	},
}

func deleteAction(projName string) {
	_, baseName := utils.GenProj(projName)
	sitePath := utils.SiteDir(baseName)

	if !utils.ProjExist(baseName) {
		utils.PrintScrn(projName+" project not found.\n", "red", 0, false)
		return
	}

	var dbName string
	if pcfg, err := utils.ReadProjectConfig(sitePath); err == nil && pcfg.DbName != "" {
		dbName = pcfg.DbName
	}

	if dbName == "" {
		wpConfigPath := filepath.Join(sitePath, "wp-config.php")
		if confData, err := os.ReadFile(wpConfigPath); err == nil {
			dbNameRegex := regexp.MustCompile(`define\(\s*['"]DB_NAME['"]\s*,\s*['"]([^'"]+)['"]\s*\);`)
			if match := dbNameRegex.FindSubmatch(confData); len(match) > 1 {
				dbName = string(match[1])
			}
		}
	}

	if dbName == "" {
		envPath := filepath.Join(sitePath, ".env")
		if envData, err := os.ReadFile(envPath); err == nil {
			envRegex := regexp.MustCompile(`(?m)^DB_DATABASE=(.+)$`)
			if match := envRegex.FindSubmatch(envData); len(match) > 1 {
				dbName = string(match[1])
			}
		}
	}

	if err := os.RemoveAll(sitePath); err != nil {
		utils.Chk_error(err, "Failed to remove project directory "+sitePath)
	}
	utils.PrintScrn("'"+baseName+"' folder deleted successfully.\n", "green", 0, false)

	if dbName != "" && utils.ChkDB(dbName) {
		utils.DeleteDBase(dbName)
		utils.PrintScrn("'"+dbName+"' database deleted successfully.\n", "green", 0, false)
	}

	utils.PrintScrn(projName+" deleted successfully.\n", "cyan", 0, false)
}
