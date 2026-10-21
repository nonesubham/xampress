package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// Supported versions
var (
	SupportedPHPVersions      = []string{"8.3", "8.2", "8.1"}
	DefaultPHPVersion         = "8.3"
	SupportedMySQLVersions    = []string{"8.0", "10.11", "10.6"}
	DefaultMySQLVersion       = "8.0"
	SupportedComposerVersions = []string{"latest", "2.7", "2.2"}
	DefaultComposerVersion    = "latest"
)

// NormalizePHPVersion returns standard major.minor version like "8.3"
func NormalizePHPVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "php")
	v = strings.TrimPrefix(v, "v")
	if strings.HasPrefix(v, "8.3") {
		return "8.3"
	}
	if strings.HasPrefix(v, "8.2") {
		return "8.2"
	}
	if strings.HasPrefix(v, "8.1") {
		return "8.1"
	}
	return v
}

// NormalizeMySQLVersion returns "8.0", "10.11", or "10.6"
func NormalizeMySQLVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "mysql")
	v = strings.TrimPrefix(v, "mariadb")
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "8") {
		return "8.0"
	}
	if strings.HasPrefix(v, "10.11") {
		return "10.11"
	}
	if strings.HasPrefix(v, "10.6") {
		return "10.6"
	}
	return v
}

// NormalizeComposerVersion returns "latest", "2.7", or "2.2"
func NormalizeComposerVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "" || strings.EqualFold(v, "latest") || strings.HasPrefix(v, "2.8") {
		return "latest"
	}
	if strings.HasPrefix(v, "2.7") {
		return "2.7"
	}
	if strings.HasPrefix(v, "2.2") {
		return "2.2"
	}
	return v
}

func GetPHPDownloadURLForVersion(ver string) (string, string, error) {
	norm := NormalizePHPVersion(ver)
	if norm == "" {
		norm = DefaultPHPVersion
	}

	var verStr string
	switch norm {
	case "8.3":
		verStr = "8.3.12"
	case "8.2":
		verStr = "8.2.24"
	case "8.1":
		verStr = "8.1.30"
	default:
		return "", "", fmt.Errorf("unsupported PHP version: %s (choose from %s)", ver, strings.Join(SupportedPHPVersions, ", "))
	}

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return fmt.Sprintf("https://dl.static-php.dev/static-php-cli/bulk/php-%s-cli-linux-x86_64.tar.gz", verStr), ".tar.gz", nil
		case "arm64":
			return fmt.Sprintf("https://dl.static-php.dev/static-php-cli/bulk/php-%s-cli-linux-aarch64.tar.gz", verStr), ".tar.gz", nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return fmt.Sprintf("https://dl.static-php.dev/static-php-cli/bulk/php-%s-cli-macos-x86_64.tar.gz", verStr), ".tar.gz", nil
		case "arm64":
			return fmt.Sprintf("https://dl.static-php.dev/static-php-cli/bulk/php-%s-cli-macos-aarch64.tar.gz", verStr), ".tar.gz", nil
		}
	case "windows":
		return fmt.Sprintf("https://dl.static-php.dev/static-php-cli/windows/spc-max/php-%s-cli-win.zip", verStr), ".zip", nil
	}
	return "", "", fmt.Errorf("unsupported OS/architecture for portable PHP: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func GetComposerDownloadURLForVersion(ver string) (string, error) {
	norm := NormalizeComposerVersion(ver)
	switch norm {
	case "latest":
		return "https://getcomposer.org/download/latest-stable/composer.phar", nil
	case "2.7":
		return "https://getcomposer.org/download/2.7.7/composer.phar", nil
	case "2.2":
		return "https://getcomposer.org/download/latest-2.2.x/composer.phar", nil
	default:
		return "", fmt.Errorf("unsupported Composer version: %s (choose from %s)", ver, strings.Join(SupportedComposerVersions, ", "))
	}
}

func GetMySQLDownloadURLForVersion(ver string) (string, string, error) {
	norm := NormalizeMySQLVersion(ver)
	if norm == "" {
		norm = DefaultMySQLVersion
	}

	switch norm {
	case "8.0":
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

	case "10.11":
		switch runtime.GOOS {
		case "linux":
			switch runtime.GOARCH {
			case "amd64":
				return "https://archive.mariadb.org/mariadb-10.11.8/bintar-linux-systemd-x86_64/mariadb-10.11.8-linux-systemd-x86_64.tar.gz", ".tar.gz", nil
			case "arm64":
				return "https://archive.mariadb.org/mariadb-10.11.8/bintar-linux-systemd-x86_64/mariadb-10.11.8-linux-systemd-x86_64.tar.gz", ".tar.gz", nil
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

	case "10.6":
		switch runtime.GOOS {
		case "linux":
			switch runtime.GOARCH {
			case "amd64":
				return "https://archive.mariadb.org/mariadb-10.6.18/bintar-linux-systemd-x86_64/mariadb-10.6.18-linux-systemd-x86_64.tar.gz", ".tar.gz", nil
			case "arm64":
				return "https://archive.mariadb.org/mariadb-10.6.18/bintar-linux-systemd-x86_64/mariadb-10.6.18-linux-systemd-x86_64.tar.gz", ".tar.gz", nil
			}
		case "darwin":
			switch runtime.GOARCH {
			case "amd64":
				return "https://cdn.mysql.com/archives/mysql-8.0/mysql-8.0.30-macos12-x86_64.tar.gz", ".tar.gz", nil
			case "arm64":
				return "https://cdn.mysql.com/archives/mysql-8.0/mysql-8.0.30-macos12-arm64.tar.gz", ".tar.gz", nil
			}
		case "windows":
			return "https://archive.mariadb.org/mariadb-10.6.18/winx64-packages/mariadb-10.6.18-winx64.zip", ".zip", nil
		}
	}

	return "", "", fmt.Errorf("unsupported MySQL version: %s (choose from %s)", ver, strings.Join(SupportedMySQLVersions, ", "))
}
