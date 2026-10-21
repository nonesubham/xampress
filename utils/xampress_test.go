package utils

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMagicLoginAndRouterWithLivePHP(t *testing.T) {
	phpBin := GetConf().PhpBin
	if !fileExists(phpBin) {
		if sysPhp, err := exec.LookPath("php"); err == nil {
			phpBin = sysPhp
		} else {
			t.Skip("PHP binary not found, skipping live PHP server test")
		}
	}

	tmpHome := t.TempDir()
	t.Setenv("XAMPRESS_HOME", tmpHome)
	sitesDir := filepath.Join(tmpHome, "sites")
	siteAlpha := filepath.Join(sitesDir, "alpha-wp")
	if err := os.MkdirAll(siteAlpha, 0755); err != nil {
		t.Fatal(err)
	}

	wpLoad := `<?php
function get_users($args) { return array((object)array('ID' => 1, 'user_login' => 'admin')); }
function get_user_by($field, $val) { return (object)array('ID' => 1, 'user_login' => 'admin'); }
function wp_clear_auth_cookie() { header('X-WP-Cleared-Cookie: true'); }
function wp_set_current_user($id, $login) { header("X-WP-Current-User: {$id}:{$login}"); }
function wp_set_auth_cookie($id, $remember) { setcookie('wordpress_logged_in_xampress', "user_{$id}_remember_" . ($remember ? '1' : '0'), time() + 3600, '/'); }
function do_action($tag, ...$args) { header("X-WP-Action: {$tag}"); }
function admin_url() { return (defined('WP_SITEURL') ? WP_SITEURL : 'http://localhost') . '/wp-admin/'; }
function wp_safe_redirect($url) { header("Location: {$url}", true, 302); }
`
	_ = os.WriteFile(filepath.Join(siteAlpha, "wp-load.php"), []byte(wpLoad), 0644)
	_ = os.WriteFile(filepath.Join(siteAlpha, "index.php"), []byte(`<?php require_once __DIR__ . '/wp-load.php'; echo "Site Alpha Home - WP_HOME=" . (defined('WP_HOME') ? WP_HOME : 'none');`), 0644)

	authPath, err := ensureAuthScript(sitesDir)
	if err != nil {
		t.Fatalf("ensureAuthScript failed: %v", err)
	}
	routerPath, err := ensureRouterScript()
	if err != nil {
		t.Fatalf("ensureRouterScript failed: %v", err)
	}
	t.Logf("Generated auth.php: %s", authPath)
	t.Logf("Generated router.php: %s", routerPath)

	cmd := exec.Command(phpBin, "-S", "127.0.0.1:18089", "-t", filepath.Join(sitesDir, "alpha-wp"), routerPath)
	cmd.Env = append(os.Environ(), "XAMPRESS_SITES_DIR="+sitesDir)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start php -S: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()
	for i := 0; i < 30; i++ {
		if isPortOpen("127.0.0.1:18089") {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 1. Test dynamic WP_HOME on http://alpha-wp.local
	req1, _ := http.NewRequest("GET", "http://127.0.0.1:18089/", nil)
	req1.Host = "alpha-wp.local"
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()
	if resp1.StatusCode != 200 || !strings.Contains(string(body1), "WP_HOME=http://alpha-wp.local") {
		t.Fatalf("unexpected home response: status=%d body=%s", resp1.StatusCode, string(body1))
	}
	t.Logf("Home request passed: %s", string(body1))

	// 2. Test Magic Login /auth.php?project=alpha-wp
	req2, _ := http.NewRequest("GET", "http://127.0.0.1:18089/auth.php?project=alpha-wp", nil)
	req2.Host = "alpha-wp.local"
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("GET /auth.php failed: %v", err)
	}
	resp2.Body.Close()

	if resp2.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 redirect from auth.php, got %d", resp2.StatusCode)
	}
	loc := resp2.Header.Get("Location")
	if loc != "http://alpha-wp.local/wp-admin/" {
		t.Fatalf("expected redirect to http://alpha-wp.local/wp-admin/, got %q", loc)
	}
	if resp2.Header.Get("X-WP-Current-User") != "1:admin" {
		t.Fatalf("expected X-WP-Current-User=1:admin, got %q", resp2.Header.Get("X-WP-Current-User"))
	}
	cookies := strings.Join(resp2.Header["Set-Cookie"], "; ")
	if !strings.Contains(cookies, "wordpress_logged_in_xampress=user_1_remember_1") {
		t.Fatalf("expected wordpress auth cookie in Set-Cookie, got %q", cookies)
	}
	t.Logf("Magic login passed! Location=%s, User=%s, Cookie=%s", loc, resp2.Header.Get("X-WP-Current-User"), cookies)

	// 3. Test invalid project /auth.php?project=missing
	req3, _ := http.NewRequest("GET", "http://127.0.0.1:18089/auth.php?project=missing", nil)
	resp3, err := client.Do(req3)
	if err != nil {
		t.Fatalf("GET /auth.php?project=missing failed: %v", err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for missing project, got %d", resp3.StatusCode)
	}
	t.Logf("Invalid project check passed with 404 Not Found")
}

func TestMDNSPacketEncodeDecode(t *testing.T) {
	txID := []byte{0x12, 0x34}
	ip := net.ParseIP("192.168.1.50").To4()
	pkt := buildMDNSResponse(txID, "alpha-wp.local", ip)

	name, nextOffset, ok := readDNSName(pkt, 12)
	if !ok || name != "alpha-wp.local" {
		t.Fatalf("expected domain alpha-wp.local, got %q (ok=%v)", name, ok)
	}
	rtype := binary.BigEndian.Uint16(pkt[nextOffset : nextOffset+2])
	rdlen := binary.BigEndian.Uint16(pkt[nextOffset+8 : nextOffset+10])
	rdata := net.IP(pkt[nextOffset+10 : nextOffset+10+int(rdlen)])
	if rtype != 1 || !bytes.Equal(rdata, ip) {
		t.Fatalf("expected A record with IP %v, got type=%d ip=%v", ip, rtype, rdata)
	}
	t.Logf("mDNS packet encode/decode passed: %s -> %s", name, rdata.String())
}

func TestLaravelProjectValidation(t *testing.T) {
	tmp := t.TempDir()

	// Should fail when xampress.json is missing
	if _, err := ValidateLaravelProject(tmp); err == nil {
		t.Fatal("expected failure when xampress.json is missing")
	}

	// Should fail when type is wordpress
	_ = WriteProjectConfig(tmp, ProjectConfig{Name: "site1", Type: ProjectTypeWordPress, DbName: "site1"})
	if _, err := ValidateLaravelProject(tmp); err == nil {
		t.Fatal("expected failure when project type is wordpress")
	}

	// Should succeed when type is laravel and artisan exists
	_ = WriteProjectConfig(tmp, ProjectConfig{Name: "site1", Type: ProjectTypeLaravel, DbName: "site1"})
	_ = os.WriteFile(filepath.Join(tmp, "artisan"), []byte("#!/usr/bin/env php\n"), 0755)
	pcfg, err := ValidateLaravelProject(tmp)
	if err != nil || pcfg.Name != "site1" {
		t.Fatalf("expected Laravel validation to pass, got err=%v", err)
	}
	t.Logf("Laravel project validation passed for %s", pcfg.Name)
}

func TestVersionNormalizationAndURLs(t *testing.T) {
	if NormalizePHPVersion("8.3.12") != "8.3" || NormalizePHPVersion("php8.2") != "8.2" || NormalizePHPVersion("8.1") != "8.1" {
		t.Errorf("unexpected NormalizePHPVersion results")
	}

	if NormalizeMySQLVersion("mysql8") != "8.0" || NormalizeMySQLVersion("10.11.8") != "10.11" || NormalizeMySQLVersion("10.6") != "10.6" {
		t.Errorf("unexpected NormalizeMySQLVersion results")
	}

	if NormalizeComposerVersion("2.7.7") != "2.7" || NormalizeComposerVersion("2.2") != "2.2" || NormalizeComposerVersion("latest") != "latest" {
		t.Errorf("unexpected NormalizeComposerVersion results")
	}

	url, ext, err := GetPHPDownloadURLForVersion("8.2")
	if err != nil || !strings.Contains(url, "8.2.24") || ext == "" {
		t.Errorf("failed GetPHPDownloadURLForVersion: url=%s ext=%s err=%v", url, ext, err)
	}

	sqlURL, sqlExt, err := GetMySQLDownloadURLForVersion("8.0")
	if err != nil || sqlURL == "" || sqlExt == "" {
		t.Errorf("failed GetMySQLDownloadURLForVersion: url=%s ext=%s err=%v", sqlURL, sqlExt, err)
	}

	compURL, err := GetComposerDownloadURLForVersion("2.7")
	if err != nil || !strings.Contains(compURL, "2.7.7") {
		t.Errorf("failed GetComposerDownloadURLForVersion: url=%s err=%v", compURL, err)
	}
}
