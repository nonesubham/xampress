package cmd

import (
	"fmt"
	"strings"
	"xampress/utils"

	"github.com/spf13/cobra"
)

var (
	serverPortFlag       int
	serverForegroundFlag bool
	serverNetworkFlag    bool
)

func init() {
	server.Flags().IntVarP(&serverPortFlag, "port", "p", 0, "Port for the PHP built-in server (auto-assigned if omitted or in use)")
	server.Flags().BoolVarP(&serverForegroundFlag, "foreground", "f", false, "Run the PHP server in the foreground")
	server.Flags().BoolVarP(&serverNetworkFlag, "network", "n", false, "Bind server to 0.0.0.0 and broadcast <site>.local over LAN mDNS")
	rootCmd.AddCommand(server)
}

var server = &cobra.Command{
	Use:   "server <start <sitename> | stop [sitename] | status>",
	Short: "Start, stop, or inspect PHP servers and <project>.local domains",
	Long: `Starts the built-in PHP server (php -S) for WordPress or Laravel projects along with
the built-in lightweight .local mDNS responder and reverse proxy. Multiple projects can
run simultaneously, each with its own <project>.local domain.

Examples:
  xampress server start my-site
  xampress server start my-second-site -n
  xampress server status
  xampress server stop my-site
  xampress server stop`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage:\n  xampress server start <sitename> [--port 8000] [-n]\n  xampress server stop [sitename]\n  xampress server status")
			return
		}

		action := strings.ToLower(args[0])
		switch action {
		case "__daemon":
			utils.RunDomainDaemon(serverNetworkFlag)
		case "start":
			if len(args) < 2 || strings.TrimSpace(args[1]) == "" {
				utils.PrintScrn("Please specify a project name to start the server.\nExample:\n  xampress server start <sitename> [--port 8000] [-n]\n", "red", 0, false)
				return
			}
			utils.StartSiteServer(args[1], serverPortFlag, serverForegroundFlag, serverNetworkFlag)
		case "stop":
			if len(args) >= 2 && strings.TrimSpace(args[1]) != "" {
				utils.StopSiteServer(args[1])
			} else {
				utils.StopAllServers()
			}
		case "status", "list":
			running := utils.GetRunningSites()
			if len(running) == 0 {
				utils.PrintScrn("No project servers are currently running.\n", "yellow", 0, false)
				return
			}
			utils.PrintScrn(fmt.Sprintf("Currently running projects (%d):\n", len(running)), "green", 0, false)
			for _, s := range running {
				netMode := "local"
				if s.Network {
					netMode = "0.0.0.0 LAN"
				}
				utils.PrintScrn(fmt.Sprintf("  • %s [%s] -> http://%s (internal port: %d, PID: %d, %s)\n", s.Site, utils.TitleCase(s.Type), s.Domain, s.Port, s.PhpPID, netMode), "cyan", 0, false)
			}
		default:
			fmt.Println("Invalid server action. Use 'start <sitename>', 'stop [sitename]', or 'status'.")
		}
	},
}
