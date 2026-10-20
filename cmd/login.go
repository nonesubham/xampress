package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"xampress/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginWP)
}

var loginWP = &cobra.Command{
	Use:   "login <projectname>",
	Short: "One-click magic login to a WordPress site's admin dashboard",
	Long: `Generates and opens the magic login URL (/auth.php?project=<projectname>) for a WordPress project,
which loads WordPress core, creates an administrator auth session cookie, and redirects to /wp-admin.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage:\n  xampress login <projectname>")
			return
		}

		_, baseFldr := utils.GenProj(args[0])
		if !utils.ProjExist(baseFldr) {
			utils.PrintScrn(fmt.Sprintf("Project '%s' not found.\n", args[0]), "red", 0, false)
			os.Exit(1)
		}

		sitePath := utils.SiteDir(baseFldr)
		if utils.DetectProjectType(sitePath) != utils.ProjectTypeWordPress {
			utils.PrintScrn(fmt.Sprintf("Project '%s' is not a WordPress site.\n", baseFldr), "red", 0, false)
			os.Exit(1)
		}

		running := utils.GetRunningSites()
		srv, isRunning := running[baseFldr]
		if !isRunning {
			utils.PrintScrn(fmt.Sprintf("Starting server for '%s'...\n", baseFldr), "cyan", 0, false)
			utils.StartSiteServer(baseFldr, 0, false, false)
			running = utils.GetRunningSites()
			srv = running[baseFldr]
		}

		magicURL := fmt.Sprintf("http://%s/auth.php?project=%s", srv.Domain, baseFldr)
		directMagicURL := fmt.Sprintf("http://localhost:%d/auth.php?project=%s", srv.Port, baseFldr)

		utils.PrintScrn(fmt.Sprintf("\nMagic Login URL for '%s':\n", baseFldr), "green", 0, false)
		utils.PrintScrn(fmt.Sprintf("  • %s\n", magicURL), "magenta", 0, false)
		utils.PrintScrn(fmt.Sprintf("  • %s (direct port)\n\n", directMagicURL), "cyan", 0, false)

		openBrowser(magicURL)
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
