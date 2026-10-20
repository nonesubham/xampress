package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"xampress/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createWP)
}

var createWP = &cobra.Command{
	Use:   "create <projectname>",
	Short: "Create a new local WordPress or Laravel project",
	Long:  `Interactively creates a new WordPress or Laravel project, sets up its MySQL database, and writes xampress.json.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			if err := utils.EnsureMySQLRunning(); err != nil {
				utils.Chk_error(err, err.Error())
			}
			projType := promptProjectType()
			if projType == utils.ProjectTypeLaravel {
				utils.CreateLaravelProject(args[0])
			} else {
				createWordPressAction(args[0])
			}
		} else {
			fmt.Println("Please enter a project name.\nExample:\n  xampress create <ProjectName>")
		}
	},
}

func promptProjectType() string {
	reader := bufio.NewReader(os.Stdin)
	utils.PrintScrn("\nSelect project type:\n", "cyan", 0, false)
	utils.PrintScrn("  1. WordPress\n", "green", 0, false)
	utils.PrintScrn("  2. Laravel\n", "magenta", 0, false)

	for {
		utils.PrintScrn("Enter choice [1/2] (default: 1): ", "cyan", 0, false)
		inp, _ := reader.ReadString('\n')
		inp = strings.TrimSpace(inp)
		if inp == "" || inp == "1" || strings.EqualFold(inp, "wordpress") || strings.EqualFold(inp, "wp") {
			return utils.ProjectTypeWordPress
		}
		if inp == "2" || strings.EqualFold(inp, "laravel") {
			return utils.ProjectTypeLaravel
		}
		utils.PrintScrn("Invalid choice. Please enter 1 (WordPress) or 2 (Laravel).\n", "red", 0, false)
	}
}

func createWordPressAction(projName string) {
	dbName, fldrName := utils.GenProj(projName)
	if utils.ChkDB(dbName) {
		utils.PrintScrn("'"+dbName+"' database already exists.\n", "red", 0, false)
		os.Exit(1)
	}
	utils.PrintScrn("Creating '"+fldrName+"' folder & downloading latest WordPress release...\n", "cyan", 0, false)
	utils.WPDown(projName)
	utils.PrintScrn("Creating '"+dbName+"' database for site...\n", "cyan", 0, false)
	utils.CreateDBase(dbName)
	utils.PrintScrn("Configuring wp-config.php and installing WordPress...\n", "cyan", 0, false)
	utils.WPInstall(projName)
}
