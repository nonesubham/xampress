package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
)

type Config struct {
	Mode         string `json:"mode"`
	Db_user      string `json:"db_user"`
	Db_pass      string `json:"db_pass"`
	Db_host      string `json:"db_host"`
	Db_port      int    `json:"db_port"`
	Wp_user      string `json:"wp_user"`
	Wp_pass      string `json:"wp_pass"`
	Wp_email     string `json:"wp_email"`
	Xampp        string `json:"xampp,omitempty"`
	SitesDir     string `json:"sites_dir"`
	PhpBin       string `json:"php_bin"`
	MysqlBin     string `json:"mysql_bin"`
	MysqldBin    string `json:"mysqld_bin,omitempty"`
	MysqlDataDir string `json:"mysql_data_dir,omitempty"`
	WpCliPath    string `json:"wp_cli_path"`
	ComposerPath string `json:"composer_path,omitempty"`
	ServerPort   int    `json:"server_port"`
}

// GetConf reads ~/.xampress/config.json and populates sensible defaults if fields are missing.
func GetConf() Config {
	homeDir := GetXampressHome()
	file, err := os.ReadFile(GetConfigPath())
	var cfg Config
	if err == nil {
		_ = json.Unmarshal(file, &cfg)
	}
	if cfg.Mode == "" {
		cfg.Mode = ModeStandalone
	}
	if cfg.Db_user == "" {
		cfg.Db_user = "root"
	}
	if cfg.Db_host == "" {
		cfg.Db_host = "127.0.0.1"
	}
	if cfg.Db_port == 0 {
		if cfg.Mode == ModeStandalone {
			cfg.Db_port = 3307
		} else {
			cfg.Db_port = 3306
		}
	}
	if cfg.Wp_user == "" {
		cfg.Wp_user = "admin"
	}
	if cfg.Wp_pass == "" {
		cfg.Wp_pass = "admin"
	}
	if cfg.Wp_email == "" {
		cfg.Wp_email = "admin@example.com"
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = 8000
	}
	if cfg.WpCliPath == "" {
		cfg.WpCliPath = filepath.Join(homeDir, "bin", "wp-cli.phar")
	}
	if cfg.ComposerPath == "" {
		cfg.ComposerPath = filepath.Join(homeDir, "bin", "composer.phar")
	}
	if cfg.Mode == ModeXAMPP {
		if cfg.Xampp == "" {
			cfg.Xampp = DefaultXamppDir()
		}
		if cfg.SitesDir == "" {
			cfg.SitesDir = filepath.Join(cfg.Xampp, "htdocs")
		}
		if cfg.PhpBin == "" {
			cfg.PhpBin = resolveXamppPhp(cfg.Xampp)
		}
		if cfg.MysqlBin == "" {
			cfg.MysqlBin = resolveXamppMysql(cfg.Xampp)
		}
	} else {
		if cfg.SitesDir == "" {
			cfg.SitesDir = filepath.Join(homeDir, "sites")
		}
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
	}
	return cfg
}

// SaveConf writes the configuration to ~/.xampress/config.json.
func SaveConf(cfg Config) error {
	if err := os.MkdirAll(GetXampressHome(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetConfigPath(), data, 0644)
}

var nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)

// GenProj normalizes a project name into a safe database name and folder name.
func GenProj(projName string) (string, string) {
	lower := strings.ToLower(strings.TrimSpace(projName))
	baseFldr := strings.Trim(nonAlphaNumRegex.ReplaceAllString(lower, "-"), "-")
	baseDb := strings.Trim(nonAlphaNumRegex.ReplaceAllString(lower, "_"), "_")
	if baseFldr == "" {
		baseFldr = "wp-site"
	}
	if baseDb == "" {
		baseDb = "wp_site"
	}
	return baseDb, baseFldr
}

// SiteDir returns the full path to a site's directory.
func SiteDir(folderName string) string {
	return filepath.Join(GetConf().SitesDir, folderName)
}

// TitleCase converts space/hyphen-separated words to Title Case without deprecated strings.Title.
func TitleCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "-", " "))
	for i, w := range words {
		runes := []rune(w)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// CheckConnection checks whether the system has internet connectivity.
func CheckConnection() bool {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// WPDown downloads WordPress core into the site's folder using WP-CLI.
func WPDown(projName string) bool {
	_, baseFldr := GenProj(projName)
	if ProjExist(baseFldr) {
		PrintScrn("'"+projName+"' site already exists, please use another name...\n", "red", 0, false)
		os.Exit(1)
	}
	sitePath := SiteDir(baseFldr)
	ferr := os.MkdirAll(sitePath, 0775)
	Chk_error(ferr, "Error creating project folder "+sitePath)

	cfg := GetConf()
	downCore := exec.Command(cfg.PhpBin, cfg.WpCliPath, "core", "download", "--path="+sitePath)
	out, err := downCore.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(sitePath)
		Chk_error(err, "Error executing WP-CLI core download: "+strings.TrimSpace(string(out)))
	}
	return true
}

// PrintScrn prints colored text to stdout.
func PrintScrn(usr_text string, tcolor string, prev_len int, clr_ovr bool) {
	colorsMap := map[string]color.Attribute{
		"green":   color.FgGreen,
		"red":     color.FgRed,
		"blue":    color.FgBlue,
		"yellow":  color.FgYellow,
		"cyan":    color.FgCyan,
		"magenta": color.FgMagenta,
		"white":   color.FgWhite,
	}

	main_clr := color.FgWhite
	if cAttr, ok := colorsMap[tcolor]; ok {
		main_clr = cAttr
	}

	if clr_ovr {
		clearScreen()
	}

	c := color.New(main_clr)
	c.Print(usr_text, strings.Repeat(" ", prev_len))
}

// clearScreen clears the terminal screen.
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// Chk_error prints an error message and exits if stus is non-nil.
func Chk_error(stus error, msg string) {
	if stus != nil {
		PrintScrn("\n"+msg+"\n", "red", 0, false)
		os.Exit(1)
	}
}

// ProjExist checks whether a project folder exists in SitesDir.
func ProjExist(folderName string) bool {
	checkPath := SiteDir(folderName)
	st, err := os.Stat(checkPath)
	return err == nil && st.IsDir()
}
