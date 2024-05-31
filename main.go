package main

import (
	"os"
	"xampress/cmd"
	"xampress/utils"
)

func main() {

	if !utils.Chk_dir() {
		utils.PrintScrn("XamPress is not inside XAMP/LAMP directory, please follow installation guide. https://github.com/nonesubham/xampress/readMe.md\n", "red", 0, true)
		os.Exit(1)
		return
	}

	if !utils.Chk_wpcli() {
		utils.PrintScrn("WP-Cli not found, Please wait downloading from https://github.com/wp-cli/wp-cli\n", "blue", 0, true)

		if !utils.Down_wpCli() {
			utils.PrintScrn("Encountering error while downloading WP-Cli\n", "red", 0, true)
			os.Exit(1)
			return
		}

		utils.PrintScrn("WP-Cli downloaded successfully, run XamPress again", "blue", 0, true)
	} else {
		cmd.Execute()
	}
}
