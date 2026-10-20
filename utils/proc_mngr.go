package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ServerState struct {
	Site    string `json:"site"`
	SiteDir string `json:"site_dir"`
	Port    int    `json:"port"`
	PhpPID  int    `json:"php_pid"`
	URL     string `json:"url"`
}

func runDir() string {
	d := filepath.Join(GetXampressHome(), "run")
	_ = os.MkdirAll(d, 0755)
	return d
}

func phpServerStatePath() string {
	return filepath.Join(runDir(), "php-server.json")
}

func mysqlPidFilePath() string {
	return filepath.Join(runDir(), "mysqld.pid")
}

// EnsureMySQLRunning starts the MySQL server (Standalone or XAMPP) if it is not already running.
func EnsureMySQLRunning() error {
	if IsMySQLRunning() {
		return nil
	}

	cfg := GetConf()
	if cfg.Mode == ModeXAMPP {
		PrintScrn("Starting XAMPP database service...\n", "cyan", 0, false)
		if err := startXamppServices(cfg.Xampp); err != nil {
			return err
		}
	} else {
		PrintScrn("Starting standalone MySQL server...\n", "cyan", 0, false)
		if err := startStandaloneMySQL(cfg); err != nil {
			return err
		}
	}

	for i := 0; i < 40; i++ {
		if IsMySQLRunning() {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("MySQL server did not become ready on %s:%d (check %s)", cfg.Db_host, cfg.Db_port, filepath.Join(runDir(), "mysqld.err"))
}

func startStandaloneMySQL(cfg Config) error {
	homeDir := GetXampressHome()
	mysqlBaseDir := filepath.Join(homeDir, "bin", "mysql")
	pidFile := mysqlPidFilePath()
	errLog := filepath.Join(runDir(), "mysqld.err")

	args := []string{
		"--no-defaults",
		"--basedir=" + mysqlBaseDir,
		"--datadir=" + cfg.MysqlDataDir,
		fmt.Sprintf("--port=%d", cfg.Db_port),
		"--bind-address=" + cfg.Db_host,
		"--pid-file=" + pidFile,
		"--log-error=" + errLog,
	}
	if runtime.GOOS != "windows" {
		sockPath := filepath.Join(runDir(), "mysqld.sock")
		args = append(args, "--socket="+sockPath)
	}

	cmd := exec.Command(cfg.MysqldBin, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start portable mysqld: %w", err)
	}
	_ = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	_ = cmd.Process.Release()
	return nil
}

func startXamppServices(xamppDir string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(filepath.Join(xamppDir, "xampp_start.exe"))
	case "darwin":
		cmd = exec.Command(filepath.Join(xamppDir, "xamppfiles", "xampp"), "start")
	default:
		cmd = exec.Command(filepath.Join(xamppDir, "lampp"), "start")
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start XAMPP: %w", err)
	}
	_ = cmd.Process.Release()
	return nil
}

// StartSiteServer starts the built-in PHP server (`php -S`) for a specific WordPress or Laravel site.
func StartSiteServer(projName string, port int, foreground bool, exposeNetwork bool) {
	_, baseFldr := GenProj(projName)
	if !ProjExist(baseFldr) {
		PrintScrn(fmt.Sprintf("Project '%s' not found in %s.\nRun `xampress list` to see available projects or `xampress create %s` to create it.\n", projName, GetConf().SitesDir, projName), "red", 0, false)
		os.Exit(1)
	}

	cfg := GetConf()
	if port <= 0 {
		port = cfg.ServerPort
	}

	if err := EnsureMySQLRunning(); err != nil {
		Chk_error(err, err.Error())
	}

	// Stop existing PHP server if one is already running
	_ = stopPHPServerOnly()

	sitePath := SiteDir(baseFldr)
	projType := DetectProjectType(sitePath)
	if projType == "" {
		PrintScrn(fmt.Sprintf("Project '%s' does not have a valid %s or wp-config.php.\n", baseFldr, ProjectConfigFile), "red", 0, false)
		os.Exit(1)
	}

	bindHost := "localhost"
	if exposeNetwork {
		bindHost = "0.0.0.0"
	}
	bindAddr := fmt.Sprintf("%s:%d", bindHost, port)
	checkAddr := fmt.Sprintf("127.0.0.1:%d", port)
	localURL := fmt.Sprintf("http://localhost:%d", port)
	networkURL := fmt.Sprintf("http://%s:%d", getLocalIP(), port)

	activeURL := localURL
	if exposeNetwork {
		activeURL = networkURL
	}

	var docRoot string
	var routerPath string
	var err error

	if projType == ProjectTypeLaravel {
		if _, vErr := ValidateLaravelProject(sitePath); vErr != nil {
			Chk_error(vErr, vErr.Error())
		}
		UpdateLaravelAppURL(sitePath, activeURL)
		docRoot = filepath.Join(sitePath, "public")
		routerPath, err = ensureLaravelRouterScript()
		Chk_error(err, "Failed to create Laravel PHP router script")
	} else {
		UpdateSiteURL(baseFldr, activeURL)
		docRoot = sitePath
		routerPath, err = ensureRouterScript()
		Chk_error(err, "Failed to create WordPress PHP router script")
	}

	logFile, err := os.OpenFile(filepath.Join(runDir(), "php-server.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	Chk_error(err, "Failed to open PHP server log file")

	cmd := exec.Command(cfg.PhpBin, "-d", "upload_max_filesize=64M", "-d", "post_max_size=64M", "-d", "memory_limit=256M", "-S", bindAddr, "-t", docRoot, routerPath)
	cmd.Dir = docRoot
	if foreground {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		Chk_error(err, fmt.Sprintf("Failed to start PHP server on %s", bindAddr))
	}

	state := ServerState{
		Site:    baseFldr,
		SiteDir: sitePath,
		Port:    port,
		PhpPID:  cmd.Process.Pid,
		URL:     localURL,
	}
	stateBytes, _ := json.MarshalIndent(state, "", "  ")
	_ = os.WriteFile(phpServerStatePath(), stateBytes, 0644)

	// Wait briefly and check if port is listening
	time.Sleep(350 * time.Millisecond)
	if !isPortOpen(checkAddr) {
		logFile.Close()
		PrintScrn(fmt.Sprintf("Failed to bind PHP server on %s (port may be in use).\n", bindAddr), "red", 0, false)
		os.Exit(1)
	}

	PrintScrn(fmt.Sprintf("\n%s project '%s' is now running on %s!\n", TitleCase(projType), baseFldr, bindAddr), "green", 0, false)
	PrintScrn(fmt.Sprintf("  Local URL     : %s\n", localURL), "cyan", 0, false)
	if projType == ProjectTypeWordPress {
		PrintScrn(fmt.Sprintf("  Dashboard URL : %s/wp-admin\n", localURL), "cyan", 0, false)
	}
	if exposeNetwork {
		PrintScrn(fmt.Sprintf("  Network URL   : %s (0.0.0.0:%d)\n", networkURL, port), "cyan", 0, false)
	}
	PrintScrn(fmt.Sprintf("  PHP PID       : %d\n", cmd.Process.Pid), "white", 0, false)

	if foreground {
		PrintScrn("\nPress Ctrl+C to stop the server...\n", "yellow", 0, false)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		PrintScrn("\nStopping PHP server...\n", "yellow", 0, false)
		_ = cmd.Process.Kill()
		_ = os.Remove(phpServerStatePath())
		logFile.Close()
		PrintScrn("Server stopped.\n", "green", 0, false)
		return
	}

	_ = cmd.Process.Release()
	logFile.Close()
	PrintScrn("  Stop Server   : xampress server stop\n\n", "yellow", 0, false)
}

// StopAllServers stops the running PHP built-in server and standalone MySQL (or XAMPP).
func StopAllServers() {
	stoppedAny := false

	if err := stopPHPServerOnly(); err == nil {
		PrintScrn("PHP server stopped successfully.\n", "green", 0, false)
		stoppedAny = true
	}

	cfg := GetConf()
	if cfg.Mode == ModeXAMPP {
		stopXamppServices(cfg.Xampp)
		PrintScrn("XAMPP services stopped.\n", "green", 0, false)
		stoppedAny = true
	} else {
		if stopStandaloneMySQL() {
			PrintScrn("Standalone MySQL server stopped.\n", "green", 0, false)
			stoppedAny = true
		}
	}

	if !stoppedAny {
		PrintScrn("No running XamPress server processes found.\n", "yellow", 0, false)
	}
}

func stopPHPServerOnly() error {
	data, err := os.ReadFile(phpServerStatePath())
	if err != nil {
		return err
	}
	var state ServerState
	if err := json.Unmarshal(data, &state); err != nil {
		_ = os.Remove(phpServerStatePath())
		return err
	}
	if state.PhpPID > 0 {
		killProcess(state.PhpPID)
	}
	_ = os.Remove(phpServerStatePath())
	return nil
}

func stopStandaloneMySQL() bool {
	pidBytes, err := os.ReadFile(mysqlPidFilePath())
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil || pid <= 0 {
		_ = os.Remove(mysqlPidFilePath())
		return false
	}
	killProcess(pid)
	_ = os.Remove(mysqlPidFilePath())
	return true
}

func stopXamppServices(xamppDir string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(filepath.Join(xamppDir, "xampp_stop.exe"))
	case "darwin":
		cmd = exec.Command(filepath.Join(xamppDir, "xamppfiles", "xampp"), "stop")
	default:
		cmd = exec.Command(filepath.Join(xamppDir, "lampp"), "stop")
	}
	_ = cmd.Run()
}

func killProcess(pid int) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	if runtime.GOOS == "windows" {
		_ = proc.Kill()
	} else {
		_ = proc.Signal(syscall.SIGTERM)
		time.Sleep(300 * time.Millisecond)
		_ = proc.Kill()
	}
}

func isPortOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "0.0.0.0"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "0.0.0.0"
}

func ensureRouterScript() (string, error) {
	routerPath := filepath.Join(runDir(), "router.php")
	routerCode := `<?php
if (!empty($_SERVER['HTTP_HOST'])) {
    $scheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';
    if (!defined('WP_HOME')) {
        define('WP_HOME', $scheme . '://' . $_SERVER['HTTP_HOST']);
    }
    if (!defined('WP_SITEURL')) {
        define('WP_SITEURL', $scheme . '://' . $_SERVER['HTTP_HOST']);
    }
}

$root = $_SERVER['DOCUMENT_ROOT'];
$path = '/' . ltrim(parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH), '/');

if ($path !== '/' && file_exists($root . $path)) {
    if (is_dir($root . $path) && substr($path, -1) !== '/') {
        header('Location: ' . $path . '/');
        exit;
    }
    if (strpos($path, '.php') === false) {
        return false;
    }
}

if (is_dir($root . $path)) {
    $index = rtrim($root . $path, '/') . '/index.php';
    if (file_exists($index)) {
        chdir(dirname($index));
        require $index;
        return true;
    }
}

if (file_exists($root . $path) && strpos($path, '.php') !== false) {
    chdir(dirname($root . $path));
    require $root . $path;
    return true;
}

chdir($root);
require $root . '/index.php';
`
	if err := os.WriteFile(routerPath, []byte(routerCode), 0644); err != nil {
		return "", err
	}
	return routerPath, nil
}

func ensureLaravelRouterScript() (string, error) {
	routerPath := filepath.Join(runDir(), "laravel-router.php")
	routerCode := `<?php
$publicPath = $_SERVER['DOCUMENT_ROOT'];
$uri = urldecode(parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH) ?? '');

if ($uri !== '/' && file_exists($publicPath . $uri)) {
    return false;
}

require_once $publicPath . '/index.php';
`
	if err := os.WriteFile(routerPath, []byte(routerCode), 0644); err != nil {
		return "", err
	}
	return routerPath, nil
}

// Start_xampp ensures the database service is running (retained for compatibility).
func Start_xampp() {
	if err := EnsureMySQLRunning(); err != nil {
		Chk_error(err, err.Error())
	}
}

// Stop_xampp stops all running servers (retained for compatibility).
func Stop_xampp() {
	StopAllServers()
}
