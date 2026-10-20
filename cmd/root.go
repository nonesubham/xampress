package cmd

import (
	"os"
	"xampress/utils"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "xampress",
		Short: "XamPress is a standalone CLI tool for managing local WordPress sites",
		Long: `XamPress automates local WordPress project management using either a built-in
Standalone environment (portable PHP, MySQL & WP-CLI in ~/.xampress) or an
existing XAMPP installation.

View this project on GitHub: https://github.com/nonesubham/xampress`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Name() == "help" {
				return
			}
			for _, a := range os.Args[1:] {
				if a == "--help" || a == "-h" {
					return
				}
			}
			// If user runs `xampress config --setup`, let config command handle it directly
			if cmd.Name() == "config" && setupFlag {
				return
			}
			if err := utils.EnsureSetup(false); err != nil {
				utils.PrintScrn(err.Error()+"\n", "red", 0, false)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
)

func init() {
	confData := utils.GetConf()
	rootCmd.AddCommand(configCMD)
	configCMD.Flags().StringVar(&sql_user, "sql-user", confData.Db_user, "Set MySQL username")
	configCMD.Flags().StringVar(&sql_pass, "sql-pass", confData.Db_pass, "Set MySQL password")
	configCMD.Flags().StringVar(&sql_host, "sql-host", confData.Db_host, "Set MySQL host")
	configCMD.Flags().IntVar(&sql_port, "sql-port", confData.Db_port, "Set MySQL port")
	configCMD.Flags().StringVar(&wp_user, "wp-user", confData.Wp_user, "Set WordPress admin username")
	configCMD.Flags().StringVar(&wp_pass, "wp-pass", confData.Wp_pass, "Set WordPress admin password")
	configCMD.Flags().StringVar(&wp_email, "wp-email", confData.Wp_email, "Set WordPress admin email")
	configCMD.Flags().StringVar(&xampp_path, "xampp-dir", confData.Xampp, "Set XAMPP path (for xampp mode)")
	configCMD.Flags().StringVar(&sites_dir, "sites-dir", confData.SitesDir, "Set WordPress sites directory")
	configCMD.Flags().StringVar(&run_mode, "mode", confData.Mode, "Set mode: 'standalone' or 'xampp'")
	configCMD.Flags().IntVar(&srv_port, "port", confData.ServerPort, "Set default PHP server port")
	configCMD.Flags().BoolVar(&setupFlag, "setup", false, "Re-run the interactive first-time setup wizard")
}

func Execute() error {
	return rootCmd.Execute()
}
