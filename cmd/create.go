package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"xampress/utils"

	"github.com/spf13/cobra"
)

var (
	createTypeFlag     string
	createVersionFlag  string
	createIsolateFlag  bool
	createPHPVerFlag   string
	createMySQLVerFlag string
	createCompVerFlag  string
)

func init() {
	createWP.Flags().StringVar(&createTypeFlag, "type", "", "Project type ('wordpress' or 'laravel')")
	createWP.Flags().StringVar(&createVersionFlag, "version", "", "WordPress or Laravel framework version (e.g. 6.6 or 11.*)")
	createWP.Flags().BoolVar(&createIsolateFlag, "isolate", false, "Isolate dedicated PHP, MySQL (for WordPress) or Composer (for Laravel) in project dir")
	createWP.Flags().StringVar(&createPHPVerFlag, "php", "", "PHP version for isolated project (e.g. 8.3, 8.2, 8.1)")
	createWP.Flags().StringVar(&createMySQLVerFlag, "mysql", "", "MySQL/MariaDB version for isolated project (e.g. 8.0, 10.11, 10.6)")
	createWP.Flags().StringVar(&createCompVerFlag, "composer", "", "Composer version for isolated Laravel project (e.g. latest, 2.7, 2.2)")
	rootCmd.AddCommand(createWP)
}

var createWP = &cobra.Command{
	Use:   "create <projectname>",
	Short: "Create a new local WordPress or Laravel project",
	Long:  `Interactively creates a new WordPress or Laravel project, sets up its MySQL database, and writes xampress.json. Supports custom framework versions and dedicated isolated runtimes via --isolate.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Please enter a project name.\nExample:\n  xampress create <ProjectName>")
			return
		}

		projName := args[0]
		_, baseFldr := utils.GenProj(projName)
		if utils.ProjExist(baseFldr) {
			utils.PrintScrn("'"+projName+"' project folder already exists, please use another name...\n", "red", 0, false)
			os.Exit(1)
		}

		projType := createTypeFlag
		if projType == "" {
			projType = promptProjectType()
		} else {
			projType = strings.ToLower(strings.TrimSpace(projType))
			if projType != utils.ProjectTypeWordPress && projType != utils.ProjectTypeLaravel {
				utils.PrintScrn("Invalid project type. Use 'wordpress' or 'laravel'.\n", "red", 0, false)
				os.Exit(1)
			}
		}

		// Optional version prompt if not passed
		frameworkVer := createVersionFlag
		reader := bufio.NewReader(os.Stdin)
		if !cmd.Flags().Changed("version") {
			if projType == utils.ProjectTypeWordPress {
				utils.PrintScrn("WordPress version [leave blank for latest]: ", "white", 0, false)
			} else {
				utils.PrintScrn("Laravel version [leave blank for latest, or e.g. 11.*, 10.*]: ", "white", 0, false)
			}
			inpVer, _ := reader.ReadString('\n')
			inpVer = strings.TrimSpace(inpVer)
			if inpVer != "" {
				frameworkVer = inpVer
			}
		}

		// If --isolate is set, user must choose versions (or use defaults / flags)
		phpVer := createPHPVerFlag
		mysqlVer := createMySQLVerFlag
		compVer := createCompVerFlag

		if createIsolateFlag {
			utils.PrintScrn("\n--- Isolated Environment Configuration ---\n", "magenta", 0, false)

			if phpVer == "" {
				defPHP := utils.DefaultPHPVersion
				utils.PrintScrn(fmt.Sprintf("Select PHP version for this project [%s] (default: %s): ", strings.Join(utils.SupportedPHPVersions, "/"), defPHP), "cyan", 0, false)
				inp, _ := reader.ReadString('\n')
				inp = utils.NormalizePHPVersion(inp)
				if inp != "" {
					phpVer = inp
				} else {
					phpVer = defPHP
				}
			}

			if projType == utils.ProjectTypeWordPress && mysqlVer == "" {
				defSQL := utils.DefaultMySQLVersion
				utils.PrintScrn(fmt.Sprintf("Select MySQL/MariaDB version for this project [%s] (default: %s): ", strings.Join(utils.SupportedMySQLVersions, "/"), defSQL), "cyan", 0, false)
				inp, _ := reader.ReadString('\n')
				inp = utils.NormalizeMySQLVersion(inp)
				if inp != "" {
					mysqlVer = inp
				} else {
					mysqlVer = defSQL
				}
			}

			if projType == utils.ProjectTypeLaravel && compVer == "" {
				defComp := utils.DefaultComposerVersion
				utils.PrintScrn(fmt.Sprintf("Select Composer version for this project [%s] (default: %s): ", strings.Join(utils.SupportedComposerVersions, "/"), defComp), "cyan", 0, false)
				inp, _ := reader.ReadString('\n')
				inp = utils.NormalizeComposerVersion(inp)
				if inp != "" {
					compVer = inp
				} else {
					compVer = defComp
				}
			}
		}

		sitePath := utils.SiteDir(baseFldr)
		pcfg := &utils.ProjectConfig{
			Name:         baseFldr,
			Type:         projType,
			FrameworkVer: frameworkVer,
		}

		if createIsolateFlag {
			if err := utils.SetupIsolatedEnvironment(sitePath, projType, phpVer, mysqlVer, compVer, pcfg); err != nil {
				_ = os.RemoveAll(sitePath)
				utils.Chk_error(err, "Failed to setup isolated environment: "+err.Error())
			}
		}

		if projType == utils.ProjectTypeLaravel {
			phpBin := ""
			compBin := ""
			if createIsolateFlag {
				phpBin = pcfg.PhpBin
				compBin = pcfg.ComposerBin
			}
			if err := utils.EnsureMySQLRunning(); err != nil {
				utils.Chk_error(err, err.Error())
			}
			utils.CreateLaravelProjectWithOptions(projName, frameworkVer, phpBin, compBin, createIsolateFlag, pcfg)
		} else {
			dbHost := ""
			dbPort := 0
			phpBin := ""
			if createIsolateFlag {
				dbHost = "127.0.0.1"
				dbPort = pcfg.MysqlPort
				phpBin = pcfg.PhpBin
				if _, _, err := utils.EnsureProjectMySQLRunning(sitePath); err != nil {
					utils.Chk_error(err, err.Error())
				}
			} else {
				if err := utils.EnsureMySQLRunning(); err != nil {
					utils.Chk_error(err, err.Error())
				}
				dbHost = utils.GetConf().Db_host
				dbPort = utils.GetConf().Db_port
			}
			createWordPressActionWithOptions(projName, frameworkVer, dbHost, dbPort, phpBin, createIsolateFlag, pcfg)
		}
	},
}

func promptProjectType() string {
	reader := bufio.NewReader(os.Stdin)
	utils.PrintScrn("\nSelect project type:\n", "cyan", 0, false)
	utils.PrintScrn("  1. WordPress\n", "green", 0, false)
	utils.PrintScrn("  2. Laravel\n", "magenta", 0, false)

	for {
		utils.PrintScrn("Enter choice [1/2] (default: 1): ", "cyan", 0, false)
		inp, _ := reader.ReadString('\n')
		inp = strings.TrimSpace(inp)
		if inp == "" || inp == "1" || strings.EqualFold(inp, "wordpress") || strings.EqualFold(inp, "wp") {
			return utils.ProjectTypeWordPress
		}
		if inp == "2" || strings.EqualFold(inp, "laravel") {
			return utils.ProjectTypeLaravel
		}
		utils.PrintScrn("Invalid choice. Please enter 1 (WordPress) or 2 (Laravel).\n", "red", 0, false)
	}
}

func createWordPressActionWithOptions(projName string, wpVer string, dbHost string, dbPort int, phpBin string, isolate bool, pcfg *utils.ProjectConfig) {
	dbName, fldrName := utils.GenProj(projName)
	if utils.ChkDB(dbName) {
		utils.PrintScrn("'"+dbName+"' database already exists.\n", "red", 0, false)
		os.Exit(1)
	}
	if wpVer != "" {
		utils.PrintScrn(fmt.Sprintf("Creating '%s' folder & downloading WordPress release %s...\n", fldrName, wpVer), "cyan", 0, false)
	} else {
		utils.PrintScrn("Creating '"+fldrName+"' folder & downloading latest WordPress release...\n", "cyan", 0, false)
	}
	utils.WPDownWithVersion(projName, wpVer, phpBin)
	utils.PrintScrn("Creating '"+dbName+"' database for site...\n", "cyan", 0, false)
	if isolate && dbPort != 0 {
		utils.CreateDBaseOnHost(dbName, dbHost, dbPort)
	} else {
		utils.CreateDBase(dbName)
	}
	utils.PrintScrn("Configuring wp-config.php and installing WordPress...\n", "cyan", 0, false)
	utils.WPInstallWithOptions(projName, dbHost, dbPort, phpBin, wpVer, isolate, pcfg)
}
