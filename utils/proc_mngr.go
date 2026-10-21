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
	Type    string `json:"type"`
	SiteDir string `json:"site_dir"`
	Port    int    `json:"port"`
	PhpPID  int    `json:"php_pid"`
	Domain  string `json:"domain"`
	URL     string `json:"url"`
	Network bool   `json:"network"`
}

type ServerRegistry struct {
	Sites     map[string]ServerState `json:"sites"`
	DaemonPID int                    `json:"daemon_pid,omitempty"`
}

func runDir() string {
	d := filepath.Join(GetXampressHome(), "run")
	_ = os.MkdirAll(d, 0755)
	return d
}

func serverRegistryPath() string {
	return filepath.Join(runDir(), "servers.json")
}

func mysqlPidFilePath() string {
	return filepath.Join(runDir(), "mysqld.pid")
}

func loadRegistry() ServerRegistry {
	var reg ServerRegistry
	data, err := os.ReadFile(serverRegistryPath())
	if err == nil {
		_ = json.Unmarshal(data, &reg)
	}
	if reg.Sites == nil {
		reg.Sites = make(map[string]ServerState)
	}
	// Prune dead processes
	changed := false
	for k, s := range reg.Sites {
		if !isPortOpen(fmt.Sprintf("127.0.0.1:%d", s.Port)) {
			delete(reg.Sites, k)
			changed = true
		}
	}
	if changed {
		saveRegistry(reg)
	}
	return reg
}

func saveRegistry(reg ServerRegistry) {
	if reg.Sites == nil {
		reg.Sites = make(map[string]ServerState)
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err == nil {
		_ = os.WriteFile(serverRegistryPath(), data, 0644)
	}
}

// GetRunningSites returns a map of all currently active project servers.
func GetRunningSites() map[string]ServerState {
	return loadRegistry().Sites
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

func findAvailablePort(startPort int) int {
	for p := startPort; p < startPort+200; p++ {
		if !isPortOpen(fmt.Sprintf("127.0.0.1:%d", p)) {
			return p
		}
	}
	return startPort
}

func ensureDomainDaemonRunning(reg *ServerRegistry, exposeNetwork bool) {
	if reg.DaemonPID > 0 {
		killProcess(reg.DaemonPID)
		reg.DaemonPID = 0
	}

	EnsurePort80Capability()

	exePath, err := os.Executable()
	if err != nil {
		return
	}

	args := []string{"server", "__daemon"}
	if exposeNetwork {
		args = append(args, "-n")
	}
	cmd := exec.Command(exePath, args...)
	if err := cmd.Start(); err == nil {
		reg.DaemonPID = cmd.Process.Pid
		_ = cmd.Process.Release()
	}
}

// StartSiteServer starts the built-in PHP server (`php -S`) for a specific WordPress or Laravel site.
// Multiple sites can run simultaneously, each with its own `<project>.local` domain and port.
func StartSiteServer(projName string, port int, foreground bool, exposeNetwork bool) {
	_, baseFldr := GenProj(projName)
	if !ProjExist(baseFldr) {
		PrintScrn(fmt.Sprintf("Project '%s' not found in %s.\nRun `xampress list` to see available projects or `xampress create %s` to create it.\n", projName, GetConf().SitesDir, projName), "red", 0, false)
		os.Exit(1)
	}

	cfg := GetConf()
	reg := loadRegistry()

	// If this specific site is already running, stop its old PHP process first
	if existing, ok := reg.Sites[baseFldr]; ok {
		killProcess(existing.PhpPID)
		delete(reg.Sites, baseFldr)
		saveRegistry(reg)
		time.Sleep(150 * time.Millisecond)
	}

	if port <= 0 {
		port = findAvailablePort(cfg.ServerPort)
	} else if isPortOpen(fmt.Sprintf("127.0.0.1:%d", port)) {
		PrintScrn(fmt.Sprintf("Port %d is already in use, selecting next free port...\n", port), "yellow", 0, false)
		port = findAvailablePort(port + 1)
	}

	if err := EnsureMySQLRunning(); err != nil {
		Chk_error(err, err.Error())
	}

	sitePath := SiteDir(baseFldr)
	projType := DetectProjectType(sitePath)
	if projType == "" {
		PrintScrn(fmt.Sprintf("Project '%s' does not have a valid %s or wp-config.php.\n", baseFldr, ProjectConfigFile), "red", 0, false)
		os.Exit(1)
	}

	pcfg, _ := ReadProjectConfig(sitePath)
	phpBin := cfg.PhpBin
	if pcfg != nil && pcfg.Isolate && pcfg.PhpBin != "" && fileExists(pcfg.PhpBin) {
		phpBin = pcfg.PhpBin
	}

	if pcfg != nil && pcfg.Isolate && pcfg.MysqldBin != "" {
		if _, _, err := EnsureProjectMySQLRunning(sitePath); err != nil {
			Chk_error(err, "Failed to start isolated MySQL: "+err.Error())
		}
	}

	bindHost := "localhost"
	if exposeNetwork {
		bindHost = "0.0.0.0"
	}
	bindAddr := fmt.Sprintf("%s:%d", bindHost, port)
	checkAddr := fmt.Sprintf("127.0.0.1:%d", port)
	domain := fmt.Sprintf("%s.local", baseFldr)
	domainURL := fmt.Sprintf("http://%s", domain)
	localURL := fmt.Sprintf("http://localhost:%d", port)
	networkURL := fmt.Sprintf("http://%s:%d", getLocalIP(), port)

	var docRoot string
	var routerPath string
	var err error

	if projType == ProjectTypeLaravel {
		if _, vErr := ValidateLaravelProject(sitePath); vErr != nil {
			Chk_error(vErr, vErr.Error())
		}
		UpdateLaravelAppURL(sitePath, domainURL)
		docRoot = filepath.Join(sitePath, "public")
		routerPath, err = ensureLaravelRouterScript()
		Chk_error(err, "Failed to create Laravel PHP router script")
	} else {
		UpdateSiteURL(baseFldr, domainURL)
		docRoot = sitePath
		_, err = ensureAuthScript(cfg.SitesDir)
		Chk_error(err, "Failed to create WordPress magic login auth.php script")
		routerPath, err = ensureRouterScript()
		Chk_error(err, "Failed to create WordPress PHP router script")
	}

	logFile, err := os.OpenFile(filepath.Join(runDir(), fmt.Sprintf("php-%s.log", baseFldr)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	Chk_error(err, "Failed to open PHP server log file")

	cmd := exec.Command(phpBin, "-d", "upload_max_filesize=64M", "-d", "post_max_size=64M", "-d", "memory_limit=256M", "-S", bindAddr, "-t", docRoot, routerPath)
	cmd.Dir = docRoot
	cmd.Env = append(os.Environ(), "XAMPRESS_SITES_DIR="+cfg.SitesDir)
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

	// Wait for PHP server port to begin listening (up to 5 seconds)
	bound := false
	for i := 0; i < 35; i++ {
		if isPortOpen(checkAddr) {
			bound = true
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if !bound {
		logFile.Close()
		PrintScrn(fmt.Sprintf("Failed to bind PHP server on %s.\n", bindAddr), "red", 0, false)
		os.Exit(1)
	}

	state := ServerState{
		Site:    baseFldr,
		Type:    projType,
		SiteDir: sitePath,
		Port:    port,
		PhpPID:  cmd.Process.Pid,
		Domain:  domain,
		URL:     domainURL,
		Network: exposeNetwork,
	}
	reg.Sites[baseFldr] = state

	anyNetwork := exposeNetwork
	for _, s := range reg.Sites {
		if s.Network {
			anyNetwork = true
		}
	}
	saveRegistry(reg)
	SyncHostsFile(reg)
	ensureDomainDaemonRunning(&reg, anyNetwork)
	saveRegistry(reg)

	PrintScrn(fmt.Sprintf("\n%s project '%s' is now running!\n", TitleCase(projType), baseFldr), "green", 0, false)
	PrintScrn(fmt.Sprintf("  Domain URL    : %s\n", domainURL), "cyan", 0, false)
	if projType == ProjectTypeWordPress {
		PrintScrn(fmt.Sprintf("  WP Admin      : %s/wp-admin\n", domainURL), "cyan", 0, false)
		PrintScrn(fmt.Sprintf("  Magic Login   : %s/auth.php?project=%s\n", domainURL, baseFldr), "magenta", 0, false)
	}
	PrintScrn(fmt.Sprintf("  Direct URL    : %s\n", localURL), "white", 0, false)
	if exposeNetwork {
		PrintScrn(fmt.Sprintf("  LAN / mDNS    : %s (or %s)\n", domainURL, networkURL), "cyan", 0, false)
	}
	PrintScrn(fmt.Sprintf("  PHP PID       : %d (Active sites: %d)\n", cmd.Process.Pid, len(reg.Sites)), "white", 0, false)

	if foreground {
		PrintScrn("\nPress Ctrl+C to stop this project server...\n", "yellow", 0, false)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		PrintScrn("\nStopping PHP server for "+baseFldr+"...\n", "yellow", 0, false)
		StopSiteServer(baseFldr)
		logFile.Close()
		return
	}

	_ = cmd.Process.Release()
	logFile.Close()
	PrintScrn(fmt.Sprintf("  Stop This Site: xampress server stop %s\n", baseFldr), "yellow", 0, false)
	PrintScrn("  Stop All      : xampress server stop\n\n", "yellow", 0, false)
}

// StopSiteServer stops a single running project server without stopping other active projects.
func StopSiteServer(projName string) {
	_, baseFldr := GenProj(projName)
	reg := loadRegistry()
	state, ok := reg.Sites[baseFldr]
	if !ok {
		PrintScrn(fmt.Sprintf("Project '%s' is not currently running.\n", baseFldr), "yellow", 0, false)
		return
	}

	if state.PhpPID > 0 {
		killProcess(state.PhpPID)
	}
	delete(reg.Sites, baseFldr)

	// If isolated MySQL pid file exists, kill isolated MySQL
	sitePath := SiteDir(baseFldr)
	isoPidPath := filepath.Join(sitePath, ".xampress", "run", "mysqld.pid")
	if pidBytes, err := os.ReadFile(isoPidPath); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes))); err == nil && pid > 0 {
			killProcess(pid)
			_ = os.Remove(isoPidPath)
		}
	}

	if len(reg.Sites) == 0 && reg.DaemonPID > 0 {
		killProcess(reg.DaemonPID)
		reg.DaemonPID = 0
	}
	saveRegistry(reg)
	SyncHostsFile(reg)
	PrintScrn(fmt.Sprintf("Stopped server for '%s' (%s).\n", baseFldr, state.Domain), "green", 0, false)
}

// StopAllServers stops all running PHP project servers, the .local domain daemon, and MySQL.
func StopAllServers() {
	stoppedAny := false
	reg := loadRegistry()

	for name, state := range reg.Sites {
		if state.PhpPID > 0 {
			killProcess(state.PhpPID)
		}
		PrintScrn(fmt.Sprintf("Stopped PHP server for '%s' (%s).\n", name, state.Domain), "green", 0, false)
		stoppedAny = true
	}

	if reg.DaemonPID > 0 {
		killProcess(reg.DaemonPID)
		stoppedAny = true
	}

	reg.Sites = make(map[string]ServerState)
	reg.DaemonPID = 0
	saveRegistry(reg)
	SyncHostsFile(reg)
	_ = os.Remove(serverRegistryPath())

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
		time.Sleep(250 * time.Millisecond)
		_ = proc.Kill()
	}
}

func isPortOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 400*time.Millisecond)
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

func ensureAuthScript(sitesDir string) (string, error) {
	authPath := filepath.Join(runDir(), "auth.php")
	authCode := fmt.Sprintf(`<?php
/**
 * XamPress WordPress Magic Login (auth.php)
 * Usage: /auth.php?project=<project-name>
 */
$sitesDir = getenv('XAMPRESS_SITES_DIR');
if (!$sitesDir) {
    $sitesDir = %q;
}
$defaultRoot = $_SERVER['DOCUMENT_ROOT'] ?? '';
$project = isset($_GET['project']) ? preg_replace('/[^a-zA-Z0-9_-]/', '', $_GET['project']) : '';

$wpRoot = '';
if ($project !== '' && $sitesDir !== '' && is_dir($sitesDir . '/' . $project)) {
    $wpRoot = $sitesDir . '/' . $project;
} elseif ($project === '' && $defaultRoot !== '' && file_exists($defaultRoot . '/wp-load.php')) {
    $wpRoot = $defaultRoot;
}

if ($wpRoot === '' || !file_exists($wpRoot . '/wp-load.php')) {
    http_response_code(404);
    header('Content-Type: text/html; charset=utf-8');
    echo '<h2>XamPress Magic Login</h2><p>WordPress project not found. Pass <code>?project=&lt;project-name&gt;</code> in the URL.</p>';
    exit;
}

chdir($wpRoot);
require_once $wpRoot . '/wp-load.php';

$admins = get_users(array(
    'role'    => 'administrator',
    'orderby' => 'ID',
    'order'   => 'ASC',
    'number'  => 1,
));

$user = !empty($admins) ? $admins[0] : get_user_by('id', 1);
if (!$user) {
    http_response_code(500);
    echo '<h2>XamPress Magic Login</h2><p>No administrator account found for this WordPress site.</p>';
    exit;
}

wp_clear_auth_cookie();
wp_set_current_user($user->ID, $user->user_login);
wp_set_auth_cookie($user->ID, true);
do_action('wp_login', $user->user_login, $user);

wp_safe_redirect(admin_url());
exit;
`, filepath.ToSlash(sitesDir))

	if err := os.WriteFile(authPath, []byte(authCode), 0644); err != nil {
		return "", err
	}
	// If in XAMPP mode, also place auth.php in htdocs so http://localhost/auth.php?project=<name> works directly
	if GetConf().Mode == ModeXAMPP && sitesDir != "" {
		_ = os.WriteFile(filepath.Join(sitesDir, "auth.php"), []byte(authCode), 0644)
	}
	return authPath, nil
}

func ensureRouterScript() (string, error) {
	routerPath := filepath.Join(runDir(), "router.php")
	authScriptPath := filepath.ToSlash(filepath.Join(runDir(), "auth.php"))
	routerCode := fmt.Sprintf(`<?php
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

if ($path === '/auth.php') {
    require %q;
    return true;
}

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
`, authScriptPath)
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
