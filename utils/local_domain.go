package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	hostsMarkerStart = "# >>> XAMPRESS-LOCAL-DOMAINS-START <<<"
	hostsMarkerEnd   = "# >>> XAMPRESS-LOCAL-DOMAINS-END <<<"
)

// CanBindPort80 tests whether the current process has permission to bind TCP port 80.
func CanBindPort80() bool {
	if isPortOpen("127.0.0.1:80") {
		return true
	}
	ln, err := net.Listen("tcp", "127.0.0.1:80")
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// EnsurePort80Capability grants `cap_net_bind_service=+ep` to the xampress binary on Linux
// so it can bind port 80 for port-free `http://<project>.local` URLs without running as root.
func EnsurePort80Capability() bool {
	if CanBindPort80() {
		return true
	}
	if runtime.GOOS == "linux" {
		exePath, err := os.Executable()
		if err != nil {
			return false
		}
		if realExe, err := filepath.EvalSymlinks(exePath); err == nil {
			exePath = realExe
		}
		PrintScrn("Granting port 80 binding permission for port-free http://<project>.local domains...\n", "cyan", 0, false)
		cmd := exec.Command("sudo", "setcap", "cap_net_bind_service=+ep", exePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	return CanBindPort80()
}

// RunDomainDaemon runs the background HTTP reverse proxy (port 80) and mDNS responder (UDP 5353)
// so multiple `<project>.local` domains work concurrently without a port number, Nginx, or Apache.
func RunDomainDaemon(bindNetwork bool) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start pure-Go mDNS responder on UDP 224.0.0.251:5353
	go startMDNSResponder()

	// Start HTTP reverse proxy on port 80
	bindHost := "127.0.0.1"
	if bindNetwork {
		bindHost = "0.0.0.0"
	}
	proxyAddr := fmt.Sprintf("%s:80", bindHost)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(r.Host)
		if colonIdx := strings.IndexByte(host, ':'); colonIdx != -1 {
			host = host[:colonIdx]
		}

		reg := loadRegistry()

		// Support http://localhost/auth.php?project=<name> by redirecting to http://<name>.local/auth.php?project=<name>
		// so that WordPress sets the auth session cookie on the target project's domain.
		if r.URL.Path == "/auth.php" {
			projParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("project")))
			if projParam != "" {
				if targetSite, ok := reg.Sites[projParam]; ok {
					targetHost := strings.ToLower(targetSite.Domain)
					if host != targetHost {
						http.Redirect(w, r, fmt.Sprintf("http://%s/auth.php?project=%s", targetHost, projParam), http.StatusFound)
						return
					}
				}
			}
		}

		for _, site := range reg.Sites {
			if strings.EqualFold(site.Domain, host) || strings.EqualFold(site.Site+".local", host) {
				targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", site.Port))
				if err != nil {
					http.Error(w, "Invalid backend target", http.StatusInternalServerError)
					return
				}
				proxy := httputil.NewSingleHostReverseProxy(targetURL)
				origDirector := proxy.Director
				proxy.Director = func(req *http.Request) {
					origHost := req.Host
					origDirector(req)
					req.Host = origHost
					req.Header.Set("X-Forwarded-Host", origHost)
					req.Header.Set("X-Forwarded-Proto", "http")
				}
				proxy.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<h2>XamPress Local Domain Router</h2><p>No active project found for <code>%s</code>.</p>", host)
	})

	srv := &http.Server{
		Addr:    proxyAddr,
		Handler: mux,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	<-sigCh
	_ = srv.Close()
}

// startMDNSResponder listens on Multicast UDP 224.0.0.251:5353 and answers A-record queries
// for any active `<project>.local` domain in the XamPress registry.
func startMDNSResponder() {
	mcastAddr := &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251),
		Port: 5353,
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		listenMDNSOnInterface(nil, mcastAddr)
		return
	}

	started := false
	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagMulticast) == 0 {
			continue
		}
		ifaceCopy := iface
		go listenMDNSOnInterface(&ifaceCopy, mcastAddr)
		started = true
	}
	if !started {
		listenMDNSOnInterface(nil, mcastAddr)
	}
}

func listenMDNSOnInterface(iface *net.Interface, mcastAddr *net.UDPAddr) {
	conn, err := net.ListenMulticastUDP("udp4", iface, mcastAddr)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetReadBuffer(65535)

	buf := make([]byte, 1500)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil || n < 12 {
			continue
		}
		handleMDNSPacket(conn, src, mcastAddr, buf[:n])
	}
}

func handleMDNSPacket(conn *net.UDPConn, src, mcastAddr *net.UDPAddr, pkt []byte) {
	flags := binary.BigEndian.Uint16(pkt[2:4])
	if (flags & 0x8000) != 0 {
		return
	}
	qdCount := int(binary.BigEndian.Uint16(pkt[4:6]))
	if qdCount == 0 {
		return
	}

	reg := loadRegistry()
	if len(reg.Sites) == 0 {
		return
	}

	offset := 12
	for i := 0; i < qdCount; i++ {
		qname, nextOffset, ok := readDNSName(pkt, offset)
		if !ok || nextOffset+4 > len(pkt) {
			return
		}
		qtype := binary.BigEndian.Uint16(pkt[nextOffset : nextOffset+2])
		qclass := binary.BigEndian.Uint16(pkt[nextOffset+2 : nextOffset+4])
		offset = nextOffset + 4

		if qtype != 1 && qtype != 255 {
			continue
		}

		lowerName := strings.TrimSuffix(strings.ToLower(qname), ".")
		for _, site := range reg.Sites {
			expected := strings.ToLower(site.Domain)
			if expected == "" {
				expected = strings.ToLower(site.Site) + ".local"
			}
			if lowerName == expected {
				ipStr := "127.0.0.1"
				if site.Network {
					if lanIP := getLocalIP(); lanIP != "0.0.0.0" {
						ipStr = lanIP
					}
				}
				ip := net.ParseIP(ipStr).To4()
				if ip == nil {
					continue
				}
				resp := buildMDNSResponse(pkt[:2], lowerName, ip)
				if (qclass & 0x8000) != 0 && src != nil {
					_, _ = conn.WriteToUDP(resp, src)
				}
				_, _ = conn.WriteToUDP(resp, mcastAddr)
				break
			}
		}
	}
}

func readDNSName(pkt []byte, offset int) (string, int, bool) {
	var labels []string
	curr := offset
	jumped := false
	nextOffset := offset

	for i := 0; i < 64; i++ {
		if curr >= len(pkt) {
			return "", 0, false
		}
		length := int(pkt[curr])
		if length == 0 {
			if !jumped {
				nextOffset = curr + 1
			}
			return strings.Join(labels, "."), nextOffset, true
		}
		if (length & 0xC0) == 0xC0 {
			if curr+1 >= len(pkt) {
				return "", 0, false
			}
			if !jumped {
				nextOffset = curr + 2
			}
			ptr := int(binary.BigEndian.Uint16(pkt[curr:curr+2]) & 0x3FFF)
			curr = ptr
			jumped = true
			continue
		}
		curr++
		if curr+length > len(pkt) {
			return "", 0, false
		}
		labels = append(labels, string(pkt[curr:curr+length]))
		curr += length
	}
	return "", 0, false
}

func buildMDNSResponse(txID []byte, domain string, ipv4 net.IP) []byte {
	var buf bytes.Buffer
	buf.Write(txID[:2])
	_ = binary.Write(&buf, binary.BigEndian, uint16(0x8400))
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))
	_ = binary.Write(&buf, binary.BigEndian, uint16(1))
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))

	for _, part := range strings.Split(domain, ".") {
		if part == "" {
			continue
		}
		buf.WriteByte(byte(len(part)))
		buf.WriteString(part)
	}
	buf.WriteByte(0)

	_ = binary.Write(&buf, binary.BigEndian, uint16(1))
	_ = binary.Write(&buf, binary.BigEndian, uint16(0x8001))
	_ = binary.Write(&buf, binary.BigEndian, uint32(120))
	_ = binary.Write(&buf, binary.BigEndian, uint16(4))
	buf.Write(ipv4[:4])
	return buf.Bytes()
}

func systemHostsPath() string {
	if runtime.GOOS == "windows" {
		sysRoot := os.Getenv("SystemRoot")
		if sysRoot == "" {
			sysRoot = `C:\Windows`
		}
		return filepath.Join(sysRoot, "System32", "drivers", "etc", "hosts")
	}
	return "/etc/hosts"
}

// SyncHostsFile updates the OS hosts file with active `<project>.local` entries so `.local` resolves
// instantly on the host machine even if mDNS is disabled locally.
func SyncHostsFile(reg ServerRegistry) {
	hostsPath := systemHostsPath()
	raw, err := os.ReadFile(hostsPath)
	if err != nil {
		return
	}

	content := string(raw)
	startIdx := strings.Index(content, hostsMarkerStart)
	endIdx := strings.Index(content, hostsMarkerEnd)
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		endIdx += len(hostsMarkerEnd)
		if endIdx < len(content) && content[endIdx] == '\r' {
			endIdx++
		}
		if endIdx < len(content) && content[endIdx] == '\n' {
			endIdx++
		}
		content = content[:startIdx] + content[endIdx:]
	}

	content = strings.TrimRight(content, "\r\n") + "\n"

	if len(reg.Sites) > 0 {
		var block strings.Builder
		block.WriteString(hostsMarkerStart + "\n")
		lanIP := getLocalIP()
		for _, s := range reg.Sites {
			dom := s.Domain
			if dom == "" {
				dom = s.Site + ".local"
			}
			block.WriteString(fmt.Sprintf("127.0.0.1 %s\n", dom))
			if s.Network && lanIP != "0.0.0.0" && lanIP != "127.0.0.1" {
				block.WriteString(fmt.Sprintf("%s %s\n", lanIP, dom))
			}
		}
		block.WriteString(hostsMarkerEnd + "\n")
		content += block.String()
	}

	if err := os.WriteFile(hostsPath, []byte(content), 0644); err != nil && errors.Is(err, os.ErrPermission) && runtime.GOOS != "windows" {
		cmd := exec.Command("sudo", "tee", hostsPath)
		cmd.Stdin = strings.NewReader(content)
		cmd.Stdout = nil
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}
}
