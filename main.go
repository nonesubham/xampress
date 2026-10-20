package main

import (
	"os"
	"xampress/cmd"
	"xampress/utils"
)

func main() {
	if err := cmd.Execute(); err != nil {
		utils.PrintScrn(err.Error()+"\n", "red", 0, false)
		os.Exit(1)
	}
}
