package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func wpConf(projName string) bool {
	return wpConfWithHost(projName, "", 0, "", "")
}

func wpConfWithHost(projName string, host string, port int, phpBin string, wpVer string) bool {
	baseDb, baseFldr := GenProj(projName)
	cfg := GetConf()
	sitePath := SiteDir(baseFldr)

	if host == "" {
		host = cfg.Db_host
	}
	if port == 0 {
		port = cfg.Db_port
	}
	if phpBin == "" {
		phpBin = cfg.PhpBin
	}

	dbHost := fmt.Sprintf("%s:%d", host, port)

	args := []string{
		cfg.WpCliPath,
		"config",
		"create",
		"--dbuser=" + cfg.Db_user,
		"--dbpass=" + cfg.Db_pass,
		"--dbname=" + baseDb,
		"--dbhost=" + dbHost,
		"--path=" + sitePath,
	}

	cmd := exec.Command(phpBin, args...)
	out, err := cmd.CombinedOutput()
	Chk_error(err, "Error executing WP-CLI config create: "+strings.TrimSpace(string(out)))
	return true
}

// WPInstall configures wp-config.php and runs `wp core install` for a new site.
func WPInstall(projName string) {
	WPInstallWithOptions(projName, "", 0, "", "", false, nil)
}

// WPInstallWithOptions installs WordPress with optional isolated settings or versions.
func WPInstallWithOptions(projName string, dbHost string, dbPort int, phpBin string, wpVer string, isolate bool, pcfg *ProjectConfig) {
	baseDb, baseFldr := GenProj(projName)
	cfg := GetConf()
	sitePath := SiteDir(baseFldr)

	if dbHost == "" {
		dbHost = cfg.Db_host
	}
	if dbPort == 0 {
		dbPort = cfg.Db_port
	}
	if phpBin == "" {
		phpBin = cfg.PhpBin
	}

	wpConfWithHost(projName, dbHost, dbPort, phpBin, wpVer)

	siteURL := fmt.Sprintf("http://localhost:%d", cfg.ServerPort)

	args := []string{
		cfg.WpCliPath,
		"core",
		"install",
		"--url=" + siteURL,
		"--title=" + TitleCase(projName),
		"--admin_user=" + cfg.Wp_user,
		"--admin_password=" + cfg.Wp_pass,
		"--admin_email=" + cfg.Wp_email,
		"--skip-email",
		"--path=" + sitePath,
	}

	cmd := exec.Command(phpBin, args...)
	out, err := cmd.CombinedOutput()
	Chk_error(err, "Error executing WP-CLI core install: "+strings.TrimSpace(string(out)))

	if pcfg == nil {
		pcfg = &ProjectConfig{}
	}
	pcfg.Name = baseFldr
	pcfg.Type = ProjectTypeWordPress
	pcfg.DbName = baseDb
	if wpVer != "" {
		pcfg.FrameworkVer = wpVer
	}
	_ = WriteProjectConfig(sitePath, *pcfg)

	PrintScrn(fmt.Sprintf("\n'%s' WordPress site created successfully!\n", projName), "green", 0, false)
	PrintScrn(fmt.Sprintf("  Site Directory     : %s\n", sitePath), "white", 0, false)
	PrintScrn(fmt.Sprintf("  Dashboard Username : %s\n", cfg.Wp_user), "cyan", 0, false)
	PrintScrn(fmt.Sprintf("  Dashboard Password : %s\n", cfg.Wp_pass), "cyan", 0, false)
	if isolate {
		PrintScrn(fmt.Sprintf("  Isolated PHP       : %s\n", pcfg.PhpVersion), "magenta", 0, false)
		PrintScrn(fmt.Sprintf("  Isolated MySQL     : %s (port %d)\n", pcfg.MySQLVersion, pcfg.MysqlPort), "magenta", 0, false)
	}
	PrintScrn(fmt.Sprintf("  Start PHP Server   : xampress server start %s\n\n", baseFldr), "yellow", 0, false)
}

// UpdateSiteURL updates WordPress `home` and `siteurl` options to match the active PHP server port.
func UpdateSiteURL(baseFldr string, siteURL string) {
	cfg := GetConf()
	sitePath := SiteDir(baseFldr)

	_ = exec.Command(cfg.PhpBin, cfg.WpCliPath, "option", "update", "home", siteURL, "--path="+sitePath).Run()
	_ = exec.Command(cfg.PhpBin, cfg.WpCliPath, "option", "update", "siteurl", siteURL, "--path="+sitePath).Run()
}
