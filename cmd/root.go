/*
Copyright © 2024 Subham nullsubham@gmail.com

*/
package cmd

import (
	"os"
	"github.com/spf13/cobra"

)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "xampress",
	Short: "Xampress, a tool that automates WordPress site management in XAMPP",
	Long: `Xampress a tool that automates WordPress site management in XAMPP,
enabling the creation of new sites and the installation of WordPress with a single click.
Ideal for achieving efficiency and ease in workflow.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
}


func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}


