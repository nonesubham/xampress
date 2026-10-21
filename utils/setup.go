package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	ModeStandalone = "standalone"
	ModeXAMPP      = "xampp"
	wpCliDownload  = "https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar"
)

// GetXampressHome returns the path to ~/.xampress.
func GetXampressHome() string {
	if customHome := os.Getenv("XAMPRESS_HOME"); customHome != "" {
		return customHome
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".xampress"
	}
	return filepath.Join(home, ".xampress")
}

// GetConfigPath returns the path to ~/.xampress/config.json.
func GetConfigPath() string {
	return filepath.Join(GetXampressHome(), "config.json")
}

// IsConfigured checks whether ~/.xampress/config.json exists and has a valid mode.
func IsConfigured() bool {
	cfgPath := GetConfigPath()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return false
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}
	return cfg.Mode == ModeStandalone || cfg.Mode == ModeXAMPP
}

// DefaultXamppDir returns the OS-specific default XAMPP directory.
func DefaultXamppDir() string {
	switch runtime.GOOS {
	case "windows":
		return "C:/xampp"
	case "darwin":
		return "/Applications/XAMPP"
	default:
		return "/opt/lampp"
	}
}

// EnsureSetup runs the interactive first-run wizard if XamPress is not yet configured (or if force is true).
func EnsureSetup(force bool) error {
	if !force && IsConfigured() {
		cfg := GetConf()
		return EnsureRuntimeBinaries(&cfg)
	}

	homeDir := GetXampressHome()
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", homeDir, err)
	}

	reader := bufio.NewReader(os.Stdin)

	PrintScrn("\n========================================================\n", "cyan", 0, false)
	PrintScrn("           Welcome to XamPress First-Run Setup          \n", "cyan", 0, false)
	PrintScrn("========================================================\n\n", "cyan", 0, false)
	PrintScrn("Choose how you want to run XamPress:\n", "white", 0, false)
	PrintScrn("  [1] Standalone Mode (Recommended)\n", "green", 0, false)
	PrintScrn("      Downloads portable PHP, MySQL & WP-CLI into ~/.xampress\n", "white", 0, false)
	PrintScrn("      Uses built-in PHP server (`php -S`) and stores sites in ~/.xampress/sites\n", "white", 0, false)
	PrintScrn("  [2] XAMPP Mode\n", "yellow", 0, false)
	PrintScrn("      Uses an existing local XAMPP installation\n\n", "white", 0, false)

	mode := ModeStandalone
	for {
		PrintScrn("Select mode [1/2] (default: 1): ", "cyan", 0, false)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" || input == "1" || strings.EqualFold(input, "standalone") {
			mode = ModeStandalone
			break
		}
		if input == "2" || strings.EqualFold(input, "xampp") {
			mode = ModeXAMPP
			break
		}
		PrintScrn("Invalid choice. Please enter 1 or 2.\n", "red", 0, false)
	}

	phpVer := DefaultPHPVersion
	mysqlVer := DefaultMySQLVersion
	composerVer := DefaultComposerVersion

	if mode == ModeStandalone {
		PrintScrn("\nSelect runtime versions to install:\n", "cyan", 0, false)
		PrintScrn(fmt.Sprintf("PHP Version [%s] (default: %s): ", strings.Join(SupportedPHPVersions, "/"), DefaultPHPVersion), "white", 0, false)
		inpPHP, _ := reader.ReadString('\n')
		inpPHP = NormalizePHPVersion(inpPHP)
		if inpPHP != "" {
			phpVer = inpPHP
		}

		PrintScrn(fmt.Sprintf("MySQL/MariaDB Version [%s] (default: %s): ", strings.Join(SupportedMySQLVersions, "/"), DefaultMySQLVersion), "white", 0, false)
		inpSQL, _ := reader.ReadString('\n')
		inpSQL = NormalizeMySQLVersion(inpSQL)
		if inpSQL != "" {
			mysqlVer = inpSQL
		}

		PrintScrn(fmt.Sprintf("Composer Version [%s] (default: %s): ", strings.Join(SupportedComposerVersions, "/"), DefaultComposerVersion), "white", 0, false)
		inpComp, _ := reader.ReadString('\n')
		inpComp = NormalizeComposerVersion(inpComp)
		if inpComp != "" {
			composerVer = inpComp
		}
	}

	cfg := Config{
		Mode:       mode,
		Db_user:    "root",
		Db_pass:    "",
		Db_host:    "127.0.0.1",
		Db_port:    3306,
		Wp_user:    "admin",
		Wp_pass:    "admin",
		Wp_email:   "admin@example.com",
		ServerPort: 8000,
	}

	if mode == ModeXAMPP {
		defXampp := DefaultXamppDir()
		for {
			PrintScrn(fmt.Sprintf("Enter XAMPP installation path (default: %s): ", defXampp), "cyan", 0, false)
			inp, _ := reader.ReadString('\n')
			inp = strings.TrimSpace(inp)
			if inp == "" {
				inp = defXampp
			}
			htdocs := filepath.Join(inp, "htdocs")
			if st, err := os.Stat(htdocs); err == nil && st.IsDir() {
				cfg.Xampp = inp
				cfg.SitesDir = htdocs
				cfg.PhpBin = resolveXamppPhp(inp)
				cfg.MysqlBin = resolveXamppMysql(inp)
				cfg.WpCliPath = filepath.Join(homeDir, "bin", "wp-cli.phar")
				break
			}
			PrintScrn(fmt.Sprintf("Could not find 'htdocs' inside '%s'. Please verify XAMPP is installed there.\n", inp), "red", 0, false)
		}
	} else {
		cfg.SitesDir = filepath.Join(homeDir, "sites")
		cfg.PhpBin = defaultStandalonePhpBin(homeDir)
		cfg.MysqlBin = defaultStandaloneMysqlBin(homeDir)
		cfg.MysqldBin = defaultStandaloneMysqldBin(homeDir)
		cfg.MysqlDataDir = filepath.Join(homeDir, "mysql-data")
		cfg.WpCliPath = filepath.Join(homeDir, "bin", "wp-cli.phar")
		cfg.ComposerPath = filepath.Join(homeDir, "bin", "composer.phar")
		cfg.Db_port = 3307 // Use port 3307 by default in standalone mode to avoid conflicting with system MySQL
		cfg.PhpVersion = phpVer
		cfg.MySQLVersion = mysqlVer
		cfg.ComposerVersion = composerVer
	}

	if err := os.MkdirAll(cfg.SitesDir, 0755); err != nil {
		return fmt.Errorf("failed to create sites directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, "run"), 0755); err != nil {
		return err
	}

	if err := SaveConf(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	if err := EnsureRuntimeBinaries(&cfg); err != nil {
		return err
	}

	PrintScrn("\nSetup completed successfully! Configuration saved to "+GetConfigPath()+"\n\n", "green", 0, false)
	return nil
}

// EnsureRuntimeBinaries ensures WP-CLI and (in standalone mode) portable PHP & MySQL are downloaded and initialized.
func EnsureRuntimeBinaries(cfg *Config) error {
	homeDir := GetXampressHome()
	binDir := filepath.Join(homeDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	if cfg.WpCliPath == "" {
		cfg.WpCliPath = filepath.Join(binDir, "wp-cli.phar")
	}

	if cfg.Mode == ModeStandalone {
		if cfg.PhpBin == "" {
			cfg.PhpBin = defaultStandalonePhpBin(homeDir)
		}
		if cfg.MysqlBin == "" {
			cfg.MysqlBin = defaultStandaloneMysqlBin(homeDir)
		}
		if cfg.MysqldBin == "" {
			cfg.MysqldBin = defaultStandaloneMysqldBin(homeDir)
		}
		if cfg.MysqlDataDir == "" {
			cfg.MysqlDataDir = filepath.Join(homeDir, "mysql-data")
		}
		if cfg.SitesDir == "" {
			cfg.SitesDir = filepath.Join(homeDir, "sites")
		}
		_ = os.MkdirAll(cfg.SitesDir, 0755)

		if !fileExists(cfg.PhpBin) {
			if err := InstallPortablePHP(homeDir, cfg); err != nil {
				return err
			}
		}

		if !fileExists(cfg.MysqldBin) {
			if err := InstallPortableMySQL(homeDir, cfg); err != nil {
				return err
			}
		}

		if !isMySQLDataInitialized(cfg.MysqlDataDir) {
			if err := InitializeMySQLDataDir(homeDir, cfg); err != nil {
				return err
			}
		}
	}

	if !fileExists(cfg.WpCliPath) {
		PrintScrn("\nInstalling WP-CLI...\n", "cyan", 0, false)
		if err := DownloadWithProgress(wpCliDownload, cfg.WpCliPath, "Downloading WP-CLI"); err != nil {
			return fmt.Errorf("failed to download WP-CLI: %w", err)
		}
		_ = os.Chmod(cfg.WpCliPath, 0755)
	}

	return SaveConf(*cfg)
}

// InstallPortablePHP downloads and extracts portable PHP CLI into ~/.xampress/bin/php.
func InstallPortablePHP(homeDir string, cfg *Config) error {
	ver := cfg.PhpVersion
	if ver == "" {
		ver = DefaultPHPVersion
	}
	url, archiveExt, err := GetPHPDownloadURLForVersion(ver)
	if err != nil {
		return err
	}

	phpDir := filepath.Join(homeDir, "bin", "php")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		return err
	}

	archivePath := filepath.Join(homeDir, "bin", "php-portable"+archiveExt)
	defer os.Remove(archivePath)

	PrintScrn(fmt.Sprintf("\nInstalling Portable PHP %s...\n", ver), "cyan", 0, false)
	if err := DownloadWithProgress(url, archivePath, fmt.Sprintf("Downloading PHP %s", ver)); err != nil {
		return fmt.Errorf("failed to download portable PHP: %w", err)
	}

	if err := ExtractArchive(archivePath, phpDir, 0, fmt.Sprintf("Extracting PHP %s", ver)); err != nil {
		return fmt.Errorf("failed to extract portable PHP: %w", err)
	}

	cfg.PhpBin = defaultStandalonePhpBin(homeDir)
	_ = os.Chmod(cfg.PhpBin, 0755)
	return nil
}

// InstallPortableMySQL downloads and extracts portable MySQL/MariaDB into ~/.xampress/bin/mysql.
func InstallPortableMySQL(homeDir string, cfg *Config) error {
	ver := cfg.MySQLVersion
	if ver == "" {
		ver = DefaultMySQLVersion
	}
	url, archiveExt, err := GetMySQLDownloadURLForVersion(ver)
	if err != nil {
		return err
	}

	mysqlDir := filepath.Join(homeDir, "bin", "mysql")
	if err := os.MkdirAll(mysqlDir, 0755); err != nil {
		return err
	}

	archivePath := filepath.Join(homeDir, "bin", "mysql-portable"+archiveExt)
	defer os.Remove(archivePath)

	PrintScrn(fmt.Sprintf("\nInstalling Portable MySQL/MariaDB %s Server...\n", ver), "cyan", 0, false)
	if err := DownloadWithProgress(url, archivePath, fmt.Sprintf("Downloading MySQL %s", ver)); err != nil {
		return fmt.Errorf("failed to download portable MySQL: %w", err)
	}

	if err := ExtractArchive(archivePath, mysqlDir, 1, fmt.Sprintf("Extracting MySQL %s", ver)); err != nil {
		return fmt.Errorf("failed to extract portable MySQL: %w", err)
	}

	cfg.MysqlBin = defaultStandaloneMysqlBin(homeDir)
	cfg.MysqldBin = defaultStandaloneMysqldBin(homeDir)
	_ = os.Chmod(cfg.MysqlBin, 0755)
	_ = os.Chmod(cfg.MysqldBin, 0755)
	return nil
}

// InitializeMySQLDataDir initializes the standalone MySQL/MariaDB data directory.
func InitializeMySQLDataDir(homeDir string, cfg *Config) error {
	PrintScrn("Initializing local MySQL database directory...\n", "cyan", 0, false)
	if err := os.MkdirAll(cfg.MysqlDataDir, 0755); err != nil {
		return err
	}

	mysqlBaseDir := filepath.Join(homeDir, "bin", "mysql")

	if runtime.GOOS == "windows" {
		installDbExe := filepath.Join(mysqlBaseDir, "bin", "mysql_install_db.exe")
		if fileExists(installDbExe) {
			cmd := exec.Command(installDbExe, "--datadir="+cfg.MysqlDataDir)
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to initialize MariaDB data dir: %v (%s)", err, string(out))
			}
			return nil
		}
	}

	// MySQL 8.0 initialization (--initialize-insecure creates root without password)
	cmd := exec.Command(
		cfg.MysqldBin,
		"--no-defaults",
		"--initialize-insecure",
		"--basedir="+mysqlBaseDir,
		"--datadir="+cfg.MysqlDataDir,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		// Fallback to mysql_install_db script if MariaDB tarball was used
		installScript := filepath.Join(mysqlBaseDir, "scripts", "mysql_install_db")
		if fileExists(installScript) {
			fallbackCmd := exec.Command(
				installScript,
				"--no-defaults",
				"--basedir="+mysqlBaseDir,
				"--datadir="+cfg.MysqlDataDir,
			)
			if out2, err2 := fallbackCmd.CombinedOutput(); err2 != nil {
				return fmt.Errorf("failed to initialize database: %v (%s)", err2, string(out2))
			}
			return nil
		}
		return fmt.Errorf("failed to initialize MySQL data directory: %v (%s)", err, string(out))
	}

	return nil
}

func isMySQLDataInitialized(dataDir string) bool {
	if dataDir == "" {
		return false
	}
	if st, err := os.Stat(filepath.Join(dataDir, "mysql")); err == nil && st.IsDir() {
		return true
	}
	return false
}

func getPhpDownloadURL() (string, string, error) {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "https://dl.static-php.dev/static-php-cli/bulk/php-8.3.12-cli-linux-x86_64.tar.gz", ".tar.gz", nil
		case "arm64":
			return "https://dl.static-php.dev/static-php-cli/bulk/php-8.3.12-cli-linux-aarch64.tar.gz", ".tar.gz", nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "https://dl.static-php.dev/static-php-cli/bulk/php-8.3.12-cli-macos-x86_64.tar.gz", ".tar.gz", nil
		case "arm64":
			return "https://dl.static-php.dev/static-php-cli/bulk/php-8.3.12-cli-macos-aarch64.tar.gz", ".tar.gz", nil
		}
	case "windows":
		return "https://dl.static-php.dev/static-php-cli/windows/spc-max/php-8.3.12-cli-win.zip", ".zip", nil
	}
	return "", "", fmt.Errorf("unsupported OS/architecture for portable PHP: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func getMySQLDownloadURL() (string, string, error) {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "https://cdn.mysql.com/archives/mysql-8.0/mysql-8.0.36-linux-glibc2.17-x86_64-minimal.tar.xz", ".tar.xz", nil
		case "arm64":
			return "https://cdn.mysql.com/archives/mysql-8.0/mysql-8.0.36-linux-glibc2.17-aarch64-minimal.tar.xz", ".tar.xz", nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "https://cdn.mysql.com/archives/mysql-8.0/mysql-8.0.30-macos12-x86_64.tar.gz", ".tar.gz", nil
		case "arm64":
			return "https://cdn.mysql.com/archives/mysql-8.0/mysql-8.0.30-macos12-arm64.tar.gz", ".tar.gz", nil
		}
	case "windows":
		return "https://archive.mariadb.org/mariadb-10.11.8/winx64-packages/mariadb-10.11.8-winx64.zip", ".zip", nil
	}
	return "", "", fmt.Errorf("unsupported OS/architecture for portable MySQL: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func defaultStandalonePhpBin(homeDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "bin", "php", "php.exe")
	}
	return filepath.Join(homeDir, "bin", "php", "php")
}

func defaultStandaloneMysqlBin(homeDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "bin", "mysql", "bin", "mysql.exe")
	}
	return filepath.Join(homeDir, "bin", "mysql", "bin", "mysql")
}

func defaultStandaloneMysqldBin(homeDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "bin", "mysql", "bin", "mysqld.exe")
	}
	return filepath.Join(homeDir, "bin", "mysql", "bin", "mysqld")
}

func resolveXamppPhp(xamppDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(xamppDir, "php", "php.exe")
	}
	binPhp := filepath.Join(xamppDir, "bin", "php")
	if fileExists(binPhp) {
		return binPhp
	}
	return filepath.Join(xamppDir, "php", "php")
}

func resolveXamppMysql(xamppDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(xamppDir, "mysql", "bin", "mysql.exe")
	}
	binMysql := filepath.Join(xamppDir, "bin", "mysql")
	if fileExists(binMysql) {
		return binMysql
	}
	return filepath.Join(xamppDir, "mysql", "bin", "mysql")
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
