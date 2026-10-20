package cmd

import (
	"os"
	"strings"
	"xampress/utils"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(artisanCmd)
}

var artisanCmd = &cobra.Command{
	Use:                "artisan [--site <project>] [artisan-arguments...]",
	Aliases:            []string{"art"},
	Short:              "Run Laravel Artisan commands in a validated XamPress Laravel project",
	DisableFlagParsing: true,
	Long: `Wraps Laravel's 'php artisan' CLI using XamPress's configured PHP binary.
Only runs after validating that the target directory contains a valid 'xampress.json'
project configuration file with '"type": "laravel"' and an 'artisan' script.

Examples:
  # Run inside a Laravel project directory
  xampress artisan migrate
  xampress artisan make:model Post -mcr

  # Or target a Laravel project by name from anywhere
  xampress artisan --site my-laravel-app migrate --seed`,
	Run: func(cmd *cobra.Command, args []string) {
		var siteName string
		var forwarded []string

		for i := 0; i < len(args); i++ {
			arg := args[i]
			if (arg == "--site" || arg == "-s") && i+1 < len(args) {
				siteName = args[i+1]
				i++
				continue
			}
			if strings.HasPrefix(arg, "--site=") {
				siteName = strings.TrimPrefix(arg, "--site=")
				continue
			}
			if strings.HasPrefix(arg, "-s=") {
				siteName = strings.TrimPrefix(arg, "-s=")
				continue
			}
			forwarded = append(forwarded, arg)
		}

		if len(forwarded) == 1 && (forwarded[0] == "--help" || forwarded[0] == "-h") && siteName == "" {
			if cwd, err := os.Getwd(); err == nil {
				if _, vErr := utils.ValidateLaravelProject(cwd); vErr != nil {
					_ = cmd.Help()
					return
				}
			}
		}

		var targetDir string
		if siteName != "" {
			_, baseFldr := utils.GenProj(siteName)
			targetDir = utils.SiteDir(baseFldr)
		} else {
			cwd, err := os.Getwd()
			utils.Chk_error(err, "Unable to determine current working directory")
			targetDir = cwd
		}

		utils.RunArtisan(targetDir, forwarded)
	},
}
