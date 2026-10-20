package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"xampress/utils"

	"github.com/spf13/cobra"
)

var cloneWP = &cobra.Command{
	Use:   "clone <source-site> <new-site>",
	Short: "Clone an existing WordPress or Laravel project",
	Long:  `Clones an existing WordPress or Laravel project's files and database into a new project.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("Usage:\n  xampress clone <source-site> <new-site>")
			return
		}
		if err := utils.EnsureMySQLRunning(); err != nil {
			utils.Chk_error(err, err.Error())
		}
		cloneAction(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(cloneWP)
}

func cloneAction(srcProj, dstProj string) {
	_, srcFldr := utils.GenProj(srcProj)
	dstDb, dstFldr := utils.GenProj(dstProj)

	if !utils.ProjExist(srcFldr) {
		utils.PrintScrn(fmt.Sprintf("Source project '%s' does not exist.\n", srcProj), "red", 0, false)
		os.Exit(1)
	}
	if utils.ProjExist(dstFldr) || utils.ChkDB(dstDb) {
		utils.PrintScrn(fmt.Sprintf("Target project or database '%s' already exists.\n", dstProj), "red", 0, false)
		os.Exit(1)
	}

	srcPath := utils.SiteDir(srcFldr)
	dstPath := utils.SiteDir(dstFldr)
	cfg := utils.GetConf()
	projType := utils.DetectProjectType(srcPath)

	utils.PrintScrn(fmt.Sprintf("Copying files from '%s' to '%s'...\n", srcFldr, dstFldr), "cyan", 0, false)
	if err := copyDir(srcPath, dstPath); err != nil {
		utils.Chk_error(err, "Error copying project files: "+err.Error())
	}

	utils.CreateDBase(dstDb)

	if projType == utils.ProjectTypeLaravel {
		envPath := filepath.Join(dstPath, ".env")
		if data, err := os.ReadFile(envPath); err == nil {
			reDb := regexp.MustCompile(`(?m)^DB_DATABASE=.*$`)
			updated := reDb.ReplaceAllString(string(data), "DB_DATABASE="+dstDb)
			_ = os.WriteFile(envPath, []byte(updated), 0644)
		}
		_ = utils.WriteProjectConfig(dstPath, utils.ProjectConfig{
			Name:   dstFldr,
			Type:   utils.ProjectTypeLaravel,
			DbName: dstDb,
		})
		migCmd := exec.Command(cfg.PhpBin, filepath.Join(dstPath, "artisan"), "migrate", "--force")
		migCmd.Dir = dstPath
		_ = migCmd.Run()
	} else {
		// WordPress project
		_ = os.Remove(filepath.Join(dstPath, "wp-config.php"))
		dbHost := fmt.Sprintf("%s:%d", cfg.Db_host, cfg.Db_port)
		confCmd := exec.Command(
			cfg.PhpBin,
			cfg.WpCliPath,
			"config",
			"create",
			"--dbuser="+cfg.Db_user,
			"--dbpass="+cfg.Db_pass,
			"--dbname="+dstDb,
			"--dbhost="+dbHost,
			"--path="+dstPath,
		)
		if out, err := confCmd.CombinedOutput(); err != nil {
			utils.Chk_error(err, "Error creating wp-config.php for cloned site: "+strings.TrimSpace(string(out)))
		}

		dumpFile := filepath.Join(dstPath, ".xampress-clone-dump.sql")
		defer os.Remove(dumpFile)

		utils.PrintScrn("Cloning database...\n", "cyan", 0, false)
		expCmd := exec.Command(cfg.PhpBin, cfg.WpCliPath, "db", "export", dumpFile, "--path="+srcPath)
		if out, err := expCmd.CombinedOutput(); err != nil {
			utils.Chk_error(err, "Error exporting source database: "+strings.TrimSpace(string(out)))
		}

		impCmd := exec.Command(cfg.PhpBin, cfg.WpCliPath, "db", "import", dumpFile, "--path="+dstPath)
		if out, err := impCmd.CombinedOutput(); err != nil {
			utils.Chk_error(err, "Error importing target database: "+strings.TrimSpace(string(out)))
		}
		_ = utils.WriteProjectConfig(dstPath, utils.ProjectConfig{
			Name:   dstFldr,
			Type:   utils.ProjectTypeWordPress,
			DbName: dstDb,
		})
	}

	utils.PrintScrn(fmt.Sprintf("\nProject '%s' cloned to '%s' successfully!\n", srcFldr, dstFldr), "green", 0, false)
	utils.PrintScrn(fmt.Sprintf("  Start Server : xampress server start %s\n\n", dstFldr), "yellow", 0, false)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
