/*
Copyright © 2024 Subham nullsubham@gmail.com
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path"
	"runtime"
	"xampress/helpers"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage config.yaml which is required for XamPress.",
	Long:  `Manage config.yaml which contains necessary dirs of XAMPP, default username, password & email for WordPress installation also default XAMPP's database credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
		
		} else {
			cmd.Flags().Visit(func(flag *pflag.Flag) {
				parentFlags(flag)
			})
		}
	},
}



// function for reset & path flag
func parentFlags(flag *pflag.Flag) {
	confPath, err := helpers.GetConfigPath()
	helpers.ErrHandler(err, true, "", false)
	if flag.Name == "path" {
		fmt.Println("Config file is at :", confPath)
	} else if flag.Name == "reset" {
		err := os.Remove(confPath)
		helpers.ErrHandler(err, true, "", false)
		fmt.Println("Config file reset to default config data successfully.")
	}
}

func init() {

	confPath, err := helpers.GetConfigPath()
	helpers.ErrHandler(err, true, "", false)

	if !helpers.DirExists(confPath) {
		fmt.Println("Config file not found, generating with default data.")
		genConf(confPath)
	}

	rootCmd.AddCommand(configCmd)

	// define flags
	configCmd.Flags().Bool("reset", false, "Restore default config data.")
	configCmd.Flags().Bool("path", false, "Path of generated config.yaml.")
}



func genConf(confPath string) {
	err := os.MkdirAll(path.Dir(confPath), os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create config directory:", err)
		os.Exit(1)
	}

	xamppDirs := map[string]map[string]string{
		"windows": {
			"xampp":    "c:/xampp/",
			"binaries": "",
			"htdocs":   "htdocs/",
		},
		"darwin": {
			"xampp":    "/Applications/XAMPP/",
			"binaries": "xamppfiles/",
			"htdocs":   "xamppfiles/htdocs/",
		},
		"linux": {
			"xampp":    "/opt/lampp/",
			"binaries": "bin/",
			"htdocs":   "htdocs/",
		},
	}

	secData := map[string]interface{}{
		"database": map[string]string{
			"user":     "root",
			"pass": "",
		},
		"wordpress": map[string]string{
			"user":     "root",
			"pass": "root",
			"email":    "email@domain.com",
		},
	}

	currOS := runtime.GOOS
	defDirs, ok := xamppDirs[currOS]
	if !ok {
		fmt.Println("XamPress does not support", currOS)
		os.Exit(1)
	}

	for key, value := range defDirs {
		viper.SetDefault(fmt.Sprintf("xampp.%s", key), value)
	}

	for section, data := range secData {
		for key, value := range data.(map[string]string) {
			viper.SetDefault(fmt.Sprintf("%s.%s", section, key), value)
		}
	}

	if err := viper.SafeWriteConfigAs(confPath); err != nil {
		fmt.Println("Something went wrong while creating config.yaml at", confPath)
		os.Exit(1)
	}

}
