package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProgressBar renders a terminal progress bar for downloads and installations.
type ProgressBar struct {
	Total      int64
	Current    int64
	Label      string
	Width      int
	StartTime  time.Time
	lastRender time.Time
}

// NewProgressBar creates and initializes a new ProgressBar.
func NewProgressBar(total int64, label string) *ProgressBar {
	return &ProgressBar{
		Total:     total,
		Label:     label,
		Width:     32,
		StartTime: time.Now(),
	}
}

// Write implements io.Writer so ProgressBar can be used with io.TeeReader or io.Copy.
func (pb *ProgressBar) Write(p []byte) (int, error) {
	n := len(p)
	pb.Current += int64(n)
	now := time.Now()
	if now.Sub(pb.lastRender) >= 80*time.Millisecond || pb.Current == pb.Total {
		pb.Render()
		pb.lastRender = now
	}
	return n, nil
}

// Render prints the current progress bar state to stdout.
func (pb *ProgressBar) Render() {
	elapsed := time.Since(pb.StartTime).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(pb.Current) / (1024 * 1024) / elapsed
	}

	if pb.Total > 0 {
		percent := float64(pb.Current) / float64(pb.Total) * 100
		if percent > 100 {
			percent = 100
		}
		filled := int(float64(pb.Width) * (percent / 100.0))
		if filled > pb.Width {
			filled = pb.Width
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.Width-filled)
		curMB := float64(pb.Current) / (1024 * 1024)
		totMB := float64(pb.Total) / (1024 * 1024)
		fmt.Printf("\r\033[K  %-24s [%s] %6.2f%% (%5.1f/%5.1f MB, %.1f MB/s)",
			pb.Label, bar, percent, curMB, totMB, speed)
	} else {
		curMB := float64(pb.Current) / (1024 * 1024)
		fmt.Printf("\r\033[K  %-24s [%-32s] %5.1f MB (%.1f MB/s)",
			pb.Label, "Downloading...", curMB, speed)
	}
}

// Finish completes the progress bar and moves to a new line.
func (pb *ProgressBar) Finish() {
	if pb.Total > 0 && pb.Current < pb.Total {
		pb.Current = pb.Total
	}
	pb.Render()
	fmt.Println()
}

// DownloadWithProgress downloads a URL to destPath while rendering a live progress bar.
func DownloadWithProgress(url string, destPath string, label string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "XamPress-CLI/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d from %s", resp.StatusCode, url)
	}

	tmpPath := destPath + ".part"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", tmpPath, err)
	}

	pb := NewProgressBar(resp.ContentLength, label)
	_, copyErr := io.Copy(out, io.TeeReader(resp.Body, pb))
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(tmpPath)
		fmt.Println()
		return fmt.Errorf("error while downloading %s: %w", label, copyErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return closeErr
	}

	pb.Finish()
	return os.Rename(tmpPath, destPath)
}

// ExtractArchive extracts .zip, .tar.gz, .tgz, or .tar.xz archives into destDir with progress indication.
func ExtractArchive(archivePath string, destDir string, stripComponents int, label string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(archivePath, destDir, stripComponents, label)
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(archivePath, destDir, stripComponents, label)
	case strings.HasSuffix(lower, ".tar.xz"):
		return extractTarXz(archivePath, destDir, stripComponents, label)
	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}

func stripPathComponents(name string, count int) string {
	cleaned := filepath.ToSlash( filepath.Clean(name) )
	cleaned = strings.TrimPrefix(cleaned, "/")
	if count <= 0 {
		return cleaned
	}
	parts := strings.Split(cleaned, "/")
	if len(parts) <= count {
		return ""
	}
	return strings.Join(parts[count:], "/")
}

func extractZip(archivePath, destDir string, stripComponents int, label string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	total := int64(len(r.File))
	pb := NewProgressBar(total, label)
	for i, f := range r.File {
		pb.Current = int64(i + 1)
		if i%25 == 0 || pb.Current == total {
			pb.Render()
		}

		relPath := stripPathComponents(f.Name, stripComponents)
		if relPath == "" {
			continue
		}
		targetPath := filepath.Join(destDir, filepath.FromSlash(relPath))
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(targetPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		mode := f.Mode()
		if mode == 0 {
			mode = 0644
		}
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	pb.Finish()
	return nil
}

func extractTarGz(archivePath, destDir string, stripComponents int, label string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	pb := NewProgressBar(stat.Size(), label)
	gzr, err := gzip.NewReader(io.TeeReader(file, pb))
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		relPath := stripPathComponents(header.Name, stripComponents)
		if relPath == "" {
			continue
		}
		targetPath := filepath.Join(destDir, filepath.FromSlash(relPath))
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			mode := os.FileMode(header.Mode)
			if mode == 0 {
				mode = 0755
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			os.Remove(targetPath)
			_ = os.Symlink(header.Linkname, targetPath)
		}
	}
	pb.Finish()
	return nil
}

func extractTarXz(archivePath, destDir string, stripComponents int, label string) error {
	fmt.Printf("  %-24s [████████████████████████████████] Extracting...\n", label)
	args := []string{"-xf", archivePath, "-C", destDir}
	if stripComponents > 0 {
		args = append(args, fmt.Sprintf("--strip-components=%d", stripComponents))
	}
	cmd := exec.Command("tar", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar extraction failed: %v (%s)", err, string(output))
	}
	return nil
}
