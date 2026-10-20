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
	server.Flags().IntVarP(&serverPortFlag, "port", "p", 0, "Port for the PHP built-in server (default from config, e.g. 8000)")
	server.Flags().BoolVarP(&serverForegroundFlag, "foreground", "f", false, "Run the PHP server in the foreground")
	server.Flags().BoolVarP(&serverNetworkFlag, "network", "n", false, "Bind server to 0.0.0.0 to allow access from local network")
	rootCmd.AddCommand(server)
}

var server = &cobra.Command{
	Use:   "server <start <sitename> | stop>",
	Short: "Start or stop the PHP server and database for a WordPress site",
	Long: `Starts the built-in PHP server (php -S) for a specific WordPress site along with the MySQL database,
or stops all running XamPress server processes.

Examples:
  xampress server start my-site
  xampress server start my-site -n
  xampress server start my-site --port 8080 -n
  xampress server stop`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage:\n  xampress server start <sitename> [--port 8000] [-n]\n  xampress server stop")
			return
		}

		action := strings.ToLower(args[0])
		switch action {
		case "start":
			if len(args) < 2 || strings.TrimSpace(args[1]) == "" {
				utils.PrintScrn("Please specify a site name to start the server.\nExample:\n  xampress server start <sitename> [--port 8000] [-n]\n", "red", 0, false)
				return
			}
			utils.StartSiteServer(args[1], serverPortFlag, serverForegroundFlag, serverNetworkFlag)
		case "stop":
			utils.StopAllServers()
		default:
			fmt.Println("Invalid server action. Use 'start <sitename>' or 'stop'.\nExample:\n  xampress server start <sitename>\n  xampress server stop")
		}
	},
}
