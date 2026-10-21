package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// SetupIsolatedEnvironment sets up dedicated PHP, MySQL, Composer in <sitePath>/.xampress/
func SetupIsolatedEnvironment(sitePath string, projType string, phpVer string, mysqlVer string, composerVer string, pcfg *ProjectConfig) error {
	isoDir := filepath.Join(sitePath, ".xampress")
	binDir := filepath.Join(isoDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create isolated bin dir: %w", err)
	}

	phpVer = NormalizePHPVersion(phpVer)
	if phpVer == "" {
		phpVer = DefaultPHPVersion
	}
	mysqlVer = NormalizeMySQLVersion(mysqlVer)
	if mysqlVer == "" {
		mysqlVer = DefaultMySQLVersion
	}
	composerVer = NormalizeComposerVersion(composerVer)
	if composerVer == "" {
		composerVer = DefaultComposerVersion
	}

	pcfg.Isolate = true
	pcfg.PhpVersion = phpVer
	pcfg.MySQLVersion = mysqlVer

	// 1. Download isolated PHP
	phpDir := filepath.Join(binDir, "php")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		return err
	}
	phpURL, ext, err := GetPHPDownloadURLForVersion(phpVer)
	if err != nil {
		return err
	}
	archivePath := filepath.Join(binDir, "php-portable"+ext)
	defer os.Remove(archivePath)

	PrintScrn(fmt.Sprintf("\n[Isolated] Downloading dedicated PHP %s...\n", phpVer), "cyan", 0, false)
	if err := DownloadWithProgress(phpURL, archivePath, fmt.Sprintf("Isolated PHP %s", phpVer)); err != nil {
		return fmt.Errorf("failed to download isolated PHP: %w", err)
	}
	if err := ExtractArchive(archivePath, phpDir, 0, fmt.Sprintf("Extracting PHP %s", phpVer)); err != nil {
		return fmt.Errorf("failed to extract isolated PHP: %w", err)
	}

	if runtime.GOOS == "windows" {
		pcfg.PhpBin = filepath.Join(phpDir, "php.exe")
	} else {
		pcfg.PhpBin = filepath.Join(phpDir, "php")
	}
	_ = os.Chmod(pcfg.PhpBin, 0755)

	// 2. If Laravel, download isolated Composer
	if projType == ProjectTypeLaravel {
		pcfg.ComposerVersion = composerVer
		compURL, err := GetComposerDownloadURLForVersion(composerVer)
		if err != nil {
			return err
		}
		compPath := filepath.Join(binDir, "composer.phar")
		PrintScrn(fmt.Sprintf("\n[Isolated] Downloading dedicated Composer %s...\n", composerVer), "cyan", 0, false)
		if err := DownloadWithProgress(compURL, compPath, fmt.Sprintf("Isolated Composer %s", composerVer)); err != nil {
			return fmt.Errorf("failed to download isolated Composer: %w", err)
		}
		_ = os.Chmod(compPath, 0755)
		pcfg.ComposerBin = compPath
	}

	// 3. Download isolated MySQL (only if WordPress, or needed)
	if projType == ProjectTypeWordPress {
		mysqlDir := filepath.Join(binDir, "mysql")
		if err := os.MkdirAll(mysqlDir, 0755); err != nil {
			return err
		}
		sqlURL, sqlExt, err := GetMySQLDownloadURLForVersion(mysqlVer)
		if err != nil {
			return err
		}
		sqlArchivePath := filepath.Join(binDir, "mysql-portable"+sqlExt)
		defer os.Remove(sqlArchivePath)

		PrintScrn(fmt.Sprintf("\n[Isolated] Downloading dedicated MySQL %s...\n", mysqlVer), "cyan", 0, false)
		if err := DownloadWithProgress(sqlURL, sqlArchivePath, fmt.Sprintf("Isolated MySQL %s", mysqlVer)); err != nil {
			return fmt.Errorf("failed to download isolated MySQL: %w", err)
		}
		if err := ExtractArchive(sqlArchivePath, mysqlDir, 1, fmt.Sprintf("Extracting MySQL %s", mysqlVer)); err != nil {
			return fmt.Errorf("failed to extract isolated MySQL: %w", err)
		}

		if runtime.GOOS == "windows" {
			pcfg.MysqlBin = filepath.Join(mysqlDir, "bin", "mysql.exe")
			pcfg.MysqldBin = filepath.Join(mysqlDir, "bin", "mysqld.exe")
		} else {
			pcfg.MysqlBin = filepath.Join(mysqlDir, "bin", "mysql")
			pcfg.MysqldBin = filepath.Join(mysqlDir, "bin", "mysqld")
		}
		_ = os.Chmod(pcfg.MysqlBin, 0755)
		_ = os.Chmod(pcfg.MysqldBin, 0755)

		pcfg.MysqlDataDir = filepath.Join(isoDir, "mysql-data")
		pcfg.MysqlPort = findAvailablePort(3310) // isolated MySQL port

		PrintScrn("[Isolated] Initializing dedicated MySQL data directory...\n", "cyan", 0, false)
		if err := initIsolatedMySQL(mysqlDir, pcfg.MysqldBin, pcfg.MysqlDataDir); err != nil {
			return err
		}
	}

	return nil
}

func initIsolatedMySQL(mysqlBaseDir string, mysqldBin string, dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		installDbExe := filepath.Join(mysqlBaseDir, "bin", "mysql_install_db.exe")
		if fileExists(installDbExe) {
			cmd := exec.Command(installDbExe, "--datadir="+dataDir)
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to initialize isolated MariaDB data dir: %v (%s)", err, string(out))
			}
			return nil
		}
	}

	cmd := exec.Command(
		mysqldBin,
		"--no-defaults",
		"--initialize-insecure",
		"--basedir="+mysqlBaseDir,
		"--datadir="+dataDir,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		installScript := filepath.Join(mysqlBaseDir, "scripts", "mysql_install_db")
		if fileExists(installScript) {
			fallbackCmd := exec.Command(
				installScript,
				"--no-defaults",
				"--basedir="+mysqlBaseDir,
				"--datadir="+dataDir,
			)
			if out2, err2 := fallbackCmd.CombinedOutput(); err2 != nil {
				return fmt.Errorf("failed to initialize isolated database: %v (%s)", err2, string(out2))
			}
			return nil
		}
		return fmt.Errorf("failed to initialize isolated MySQL data directory: %v (%s)", err, string(out))
	}
	return nil
}

// EnsureProjectMySQLRunning starts isolated mysqld if project is isolated, or global MySQL if not.
func EnsureProjectMySQLRunning(sitePath string) (string, int, error) {
	pcfg, err := ReadProjectConfig(sitePath)
	if err == nil && pcfg != nil && pcfg.Isolate && pcfg.MysqldBin != "" {
		port := pcfg.MysqlPort
		if port == 0 {
			port = 3307
		}
		if isPortOpen(fmt.Sprintf("127.0.0.1:%d", port)) {
			return "127.0.0.1", port, nil
		}

		PrintScrn(fmt.Sprintf("Starting dedicated isolated MySQL server on port %d...\n", port), "cyan", 0, false)
		runD := filepath.Join(sitePath, ".xampress", "run")
		_ = os.MkdirAll(runD, 0755)
		pidFile := filepath.Join(runD, "mysqld.pid")
		errLog := filepath.Join(runD, "mysqld.err")
		mysqlBaseDir := filepath.Dir(filepath.Dir(pcfg.MysqldBin))

		args := []string{
			"--no-defaults",
			"--basedir=" + mysqlBaseDir,
			"--datadir=" + pcfg.MysqlDataDir,
			fmt.Sprintf("--port=%d", port),
			"--bind-address=127.0.0.1",
			"--pid-file=" + pidFile,
			"--log-error=" + errLog,
		}
		if runtime.GOOS != "windows" {
			sockPath := filepath.Join(runD, "mysqld.sock")
			args = append(args, "--socket="+sockPath)
		}

		cmd := exec.Command(pcfg.MysqldBin, args...)
		if err := cmd.Start(); err != nil {
			return "", 0, fmt.Errorf("unable to start isolated mysqld: %w", err)
		}
		_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
		_ = cmd.Process.Release()

		for i := 0; i < 40; i++ {
			if isPortOpen(fmt.Sprintf("127.0.0.1:%d", port)) {
				return "127.0.0.1", port, nil
			}
			// brief sleep
			exec.Command("sleep", "0.2").Run()
		}
		return "127.0.0.1", port, nil
	}

	// Fallback to global MySQL
	cfg := GetConf()
	if err := EnsureMySQLRunning(); err != nil {
		return "", 0, err
	}
	return cfg.Db_host, cfg.Db_port, nil
}
