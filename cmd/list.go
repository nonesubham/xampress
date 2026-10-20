package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"xampress/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listWP)
}

var listWP = &cobra.Command{
	Use:   "list",
	Short: "List all installed WordPress and Laravel projects",
	Long:  `Lists all installed WordPress and Laravel projects in the configured sites directory and shows their running status.`,
	Run: func(cmd *cobra.Command, args []string) {
		listAction()
	},
}

type siteEntry struct {
	Folder string
	Type   string
}

func listAction() {
	sitesDir := utils.GetConf().SitesDir
	_ = os.MkdirAll(sitesDir, 0755)
	folders, htErr := os.ReadDir(sitesDir)
	utils.Chk_error(htErr, "Unable to read sites directory: "+sitesDir)

	var sites []siteEntry
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		sitePath := filepath.Join(sitesDir, folder.Name())
		pType := utils.DetectProjectType(sitePath)
		if pType != "" {
			sites = append(sites, siteEntry{
				Folder: folder.Name(),
				Type:   pType,
			})
		}
	}

	running := utils.GetRunningSites()

	utils.PrintScrn(fmt.Sprintf("Total %s project(s) in %s:\n", strconv.Itoa(len(sites)), sitesDir), "yellow", 0, false)
	for i, s := range sites {
		utils.PrintScrn(fmt.Sprintf("  %d. ", i+1), "white", 0, false)
		badgeColor := "green"
		if s.Type == utils.ProjectTypeLaravel {
			badgeColor = "cyan"
		}
		utils.PrintScrn(fmt.Sprintf("[%s] ", utils.TitleCase(s.Type)), badgeColor, 0, false)
		if srv, ok := running[s.Folder]; ok {
			utils.PrintScrn(fmt.Sprintf("%s (%s) ", utils.TitleCase(s.Folder), s.Folder), "magenta", 0, false)
			utils.PrintScrn(fmt.Sprintf("● RUNNING -> http://%s\n", srv.Domain), "green", 0, false)
		} else {
			utils.PrintScrn(fmt.Sprintf("%s (http://%s.local)\n", utils.TitleCase(s.Folder), s.Folder), "magenta", 0, false)
		}
	}
}
