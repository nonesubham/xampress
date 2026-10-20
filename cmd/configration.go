package cmd

import (
	"fmt"
	"os"
	"xampress/utils"

	"github.com/spf13/cobra"
)

var (
	sql_user   string
	sql_pass   string
	sql_host   string
	sql_port   int
	wp_user    string
	wp_pass    string
	wp_email   string
	xampp_path string
	sites_dir  string
	run_mode   string
	srv_port   int
	setupFlag  bool
)

var configCMD = &cobra.Command{
	Use:   "config",
	Short: "View or update XamPress configuration, credentials, or mode",
	Long:  "Configure MySQL & WordPress credentials, default PHP server port, or switch between standalone and xampp mode.",
	Run: func(cmd *cobra.Command, args []string) {
		if setupFlag {
			if err := utils.EnsureSetup(true); err != nil {
				utils.PrintScrn(err.Error()+"\n", "red", 0, false)
				os.Exit(1)
			}
			return
		}

		cfg := utils.GetConf()
		changed := false

		if cmd.Flags().Changed("sql-user") {
			cfg.Db_user = sql_user
			changed = true
		}
		if cmd.Flags().Changed("sql-pass") {
			cfg.Db_pass = sql_pass
			changed = true
		}
		if cmd.Flags().Changed("sql-host") {
			cfg.Db_host = sql_host
			changed = true
		}
		if cmd.Flags().Changed("sql-port") {
			cfg.Db_port = sql_port
			changed = true
		}
		if cmd.Flags().Changed("wp-user") {
			cfg.Wp_user = wp_user
			changed = true
		}
		if cmd.Flags().Changed("wp-pass") {
			cfg.Wp_pass = wp_pass
			changed = true
		}
		if cmd.Flags().Changed("wp-email") {
			cfg.Wp_email = wp_email
			changed = true
		}
		if cmd.Flags().Changed("xampp-dir") {
			cfg.Xampp = xampp_path
			changed = true
		}
		if cmd.Flags().Changed("sites-dir") {
			cfg.SitesDir = sites_dir
			changed = true
		}
		if cmd.Flags().Changed("mode") {
			if run_mode != utils.ModeStandalone && run_mode != utils.ModeXAMPP {
				utils.PrintScrn("Invalid mode. Use 'standalone' or 'xampp'.\n", "red", 0, false)
				os.Exit(1)
			}
			cfg.Mode = run_mode
			changed = true
		}
		if cmd.Flags().Changed("port") {
			cfg.ServerPort = srv_port
			changed = true
		}

		if !changed {
			utils.PrintScrn("\nCurrent XamPress Configuration ("+utils.GetConfigPath()+"):\n", "cyan", 0, false)
			utils.PrintScrn(fmt.Sprintf("  Mode            : %s\n", cfg.Mode), "green", 0, false)
			utils.PrintScrn(fmt.Sprintf("  Sites Directory : %s\n", cfg.SitesDir), "white", 0, false)
			utils.PrintScrn(fmt.Sprintf("  PHP Binary      : %s\n", cfg.PhpBin), "white", 0, false)
			utils.PrintScrn(fmt.Sprintf("  WP-CLI Path     : %s\n", cfg.WpCliPath), "white", 0, false)
			utils.PrintScrn(fmt.Sprintf("  MySQL Endpoint  : %s@%s:%d\n", cfg.Db_user, cfg.Db_host, cfg.Db_port), "white", 0, false)
			utils.PrintScrn(fmt.Sprintf("  Default Port    : %d\n", cfg.ServerPort), "white", 0, false)
			utils.PrintScrn(fmt.Sprintf("  WP Admin User   : %s (%s)\n\n", cfg.Wp_user, cfg.Wp_email), "white", 0, false)
			utils.PrintScrn("Use `xampress config --help` to see flags or `xampress config --setup` to re-run setup.\n", "yellow", 0, false)
			return
		}

		err := utils.SaveConf(cfg)
		utils.Chk_error(err, "Error while updating configuration")
		if err := utils.EnsureRuntimeBinaries(&cfg); err != nil {
			utils.Chk_error(err, err.Error())
		}
		utils.PrintScrn("Configuration updated successfully in "+utils.GetConfigPath()+"\n", "green", 0, false)
	},
}
