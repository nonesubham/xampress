package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	ProjectTypeWordPress = "wordpress"
	ProjectTypeLaravel   = "laravel"
	ProjectConfigFile    = "xampress.json"
	composerDownloadURL  = "https://getcomposer.org/download/latest-stable/composer.phar"
)

// ProjectConfig is stored in `<project-dir>/xampress.json` to identify and validate XamPress projects.
type ProjectConfig struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	DbName    string `json:"db_name"`
	CreatedAt string `json:"created_at"`
}

// WriteProjectConfig writes `xampress.json` inside the given project directory.
func WriteProjectConfig(sitePath string, pcfg ProjectConfig) error {
	if pcfg.CreatedAt == "" {
		pcfg.CreatedAt = time.Now().Format(time.RFC3339)
	}
	data, err := json.MarshalIndent(pcfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sitePath, ProjectConfigFile), data, 0644)
}

// ReadProjectConfig reads and parses `xampress.json` from the given directory.
func ReadProjectConfig(dirPath string) (*ProjectConfig, error) {
	cfgPath := filepath.Join(dirPath, ProjectConfigFile)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var pcfg ProjectConfig
	if err := json.Unmarshal(data, &pcfg); err != nil {
		return nil, err
	}
	return &pcfg, nil
}

// ValidateLaravelProject verifies that dirPath contains a valid `xampress.json` with `"type": "laravel"` and an `artisan` file.
func ValidateLaravelProject(dirPath string) (*ProjectConfig, error) {
	pcfg, err := ReadProjectConfig(dirPath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: '%s' not found in %s (not a XamPress project)", ProjectConfigFile, dirPath)
	}
	if !strings.EqualFold(pcfg.Type, ProjectTypeLaravel) {
		return nil, fmt.Errorf("validation failed: project '%s' in %s has type '%s', expected 'laravel'", pcfg.Name, dirPath, pcfg.Type)
	}
	artisanPath := filepath.Join(dirPath, "artisan")
	if !fileExists(artisanPath) {
		return nil, fmt.Errorf("validation failed: 'artisan' script not found in %s", dirPath)
	}
	return pcfg, nil
}

// DetectProjectType determines if a site directory is a Laravel or WordPress project.
func DetectProjectType(sitePath string) string {
	if pcfg, err := ReadProjectConfig(sitePath); err == nil && pcfg.Type != "" {
		return strings.ToLower(pcfg.Type)
	}
	if fileExists(filepath.Join(sitePath, "wp-config.php")) {
		return ProjectTypeWordPress
	}
	return ""
}

// EnsureComposerInstalled downloads composer.phar with a progress bar if not already installed.
func EnsureComposerInstalled(cfg *Config) error {
	if cfg.ComposerPath == "" {
		cfg.ComposerPath = filepath.Join(GetXampressHome(), "bin", "composer.phar")
	}
	if fileExists(cfg.ComposerPath) {
		return nil
	}
	PrintScrn("\nInstalling Composer...\n", "cyan", 0, false)
	if err := DownloadWithProgress(composerDownloadURL, cfg.ComposerPath, "Downloading Composer"); err != nil {
		return fmt.Errorf("failed to download Composer: %w", err)
	}
	_ = os.Chmod(cfg.ComposerPath, 0755)
	return SaveConf(*cfg)
}

// CreateLaravelProject creates a new Laravel project via Composer, configures .env, writes xampress.json, and runs initial migrations.
func CreateLaravelProject(projName string) {
	dbName, baseFldr := GenProj(projName)
	if ProjExist(baseFldr) {
		PrintScrn("'"+projName+"' project folder already exists, please use another name...\n", "red", 0, false)
		os.Exit(1)
	}
	if ChkDB(dbName) {
		PrintScrn("'"+dbName+"' database already exists...\n", "red", 0, false)
		os.Exit(1)
	}

	cfg := GetConf()
	if err := EnsureComposerInstalled(&cfg); err != nil {
		Chk_error(err, err.Error())
	}

	sitePath := SiteDir(baseFldr)
	PrintScrn(fmt.Sprintf("Creating Laravel project '%s' via Composer...\n", baseFldr), "cyan", 0, false)

	composerCmd := exec.Command(cfg.PhpBin, cfg.ComposerPath, "create-project", "laravel/laravel", sitePath, "--prefer-dist")
	composerCmd.Stdout = os.Stdout
	composerCmd.Stderr = os.Stderr
	if err := composerCmd.Run(); err != nil {
		_ = os.RemoveAll(sitePath)
		Chk_error(err, "Error creating Laravel project via Composer")
	}

	PrintScrn("Creating '"+dbName+"' database for Laravel project...\n", "cyan", 0, false)
	CreateDBase(dbName)

	appURL := fmt.Sprintf("http://localhost:%d", cfg.ServerPort)
	PrintScrn("Configuring Laravel .env and generating xampress.json...\n", "cyan", 0, false)
	if err := configureLaravelEnv(sitePath, TitleCase(projName), appURL, dbName, cfg); err != nil {
		Chk_error(err, "Failed to configure Laravel .env: "+err.Error())
	}

	pcfg := ProjectConfig{
		Name:   baseFldr,
		Type:   ProjectTypeLaravel,
		DbName: dbName,
	}
	if err := WriteProjectConfig(sitePath, pcfg); err != nil {
		Chk_error(err, "Failed to write "+ProjectConfigFile+": "+err.Error())
	}

	// Run initial migrations
	PrintScrn("Running initial database migrations (`php artisan migrate`)...\n", "cyan", 0, false)
	migCmd := exec.Command(cfg.PhpBin, filepath.Join(sitePath, "artisan"), "migrate", "--force")
	migCmd.Dir = sitePath
	migCmd.Stdout = os.Stdout
	migCmd.Stderr = os.Stderr
	_ = migCmd.Run()

	PrintScrn(fmt.Sprintf("\nLaravel project '%s' created successfully!\n", projName), "green", 0, false)
	PrintScrn(fmt.Sprintf("  Project Directory : %s\n", sitePath), "white", 0, false)
	PrintScrn(fmt.Sprintf("  Project Config    : %s\n", filepath.Join(sitePath, ProjectConfigFile)), "white", 0, false)
	PrintScrn(fmt.Sprintf("  Database Name     : %s\n", dbName), "cyan", 0, false)
	PrintScrn(fmt.Sprintf("  Start Server      : xampress server start %s\n", baseFldr), "yellow", 0, false)
	PrintScrn(fmt.Sprintf("  Run Artisan (CWD) : cd %s && xampress artisan <command>\n", sitePath), "yellow", 0, false)
	PrintScrn(fmt.Sprintf("  Run Artisan       : xampress artisan --site %s <command>\n\n", baseFldr), "yellow", 0, false)
}

func configureLaravelEnv(sitePath, appName, appURL, dbName string, cfg Config) error {
	envPath := filepath.Join(sitePath, ".env")
	examplePath := filepath.Join(sitePath, ".env.example")

	if !fileExists(envPath) && fileExists(examplePath) {
		data, err := os.ReadFile(examplePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(envPath, data, 0644); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}
	content := string(data)

	replacements := map[string]string{
		"APP_NAME":      fmt.Sprintf("\"%s\"", appName),
		"APP_URL":       appURL,
		"DB_CONNECTION": "mysql",
		"DB_HOST":       cfg.Db_host,
		"DB_PORT":       fmt.Sprintf("%d", cfg.Db_port),
		"DB_DATABASE":   dbName,
		"DB_USERNAME":   cfg.Db_user,
		"DB_PASSWORD":   cfg.Db_pass,
	}

	for key, val := range replacements {
		re := regexp.MustCompile(`(?m)^#?\s*` + regexp.QuoteMeta(key) + `=.*$`)
		line := fmt.Sprintf("%s=%s", key, val)
		if re.MatchString(content) {
			content = re.ReplaceAllString(content, line)
		} else {
			content += "\n" + line
		}
	}

	return os.WriteFile(envPath, []byte(content), 0644)
}

// UpdateLaravelAppURL updates APP_URL in the Laravel project's .env file.
func UpdateLaravelAppURL(sitePath, appURL string) {
	envPath := filepath.Join(sitePath, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`(?m)^APP_URL=.*$`)
	updated := re.ReplaceAllString(string(data), "APP_URL="+appURL)
	_ = os.WriteFile(envPath, []byte(updated), 0644)
}

// RunArtisan validates that targetDir is a XamPress Laravel project via `xampress.json` and executes `php artisan`.
func RunArtisan(targetDir string, artisanArgs []string) {
	pcfg, err := ValidateLaravelProject(targetDir)
	if err != nil {
		PrintScrn("\n"+err.Error()+"\n", "red", 0, false)
		PrintScrn("Hint: Run `xampress artisan` from inside a Laravel project directory containing 'xampress.json', or pass `--site <project-name>`.\n\n", "yellow", 0, false)
		os.Exit(1)
	}

	// Ensure MySQL server is running so database commands like `migrate` or `db:seed` work seamlessly
	if err := EnsureMySQLRunning(); err != nil {
		PrintScrn("Warning: could not auto-start MySQL ("+err.Error()+")\n", "yellow", 0, false)
	}

	cfg := GetConf()
	artisanFile := filepath.Join(targetDir, "artisan")
	cmdArgs := append([]string{artisanFile}, artisanArgs...)

	PrintScrn(fmt.Sprintf("[XamPress Laravel: %s] Running: php artisan %s\n", pcfg.Name, strings.Join(artisanArgs, " ")), "cyan", 0, false)

	cmd := exec.Command(cfg.PhpBin, cmdArgs...)
	cmd.Dir = targetDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		Chk_error(err, "Failed to execute artisan command")
	}
}
