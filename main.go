package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var partPattern = regexp.MustCompile(`^(.*_)(\d+)(\.pkg)$`)

const probeBatch = 8

var probeClient = &http.Client{Timeout: 30 * time.Second}

var (
	pkgDir   string
	autoMode = -1
	seenMu   sync.Mutex
	seen     = map[string]bool{}
)

func packageKey(u *url.URL) string {
	name := path.Base(u.Path)
	if m := partPattern.FindStringSubmatch(name); m != nil {
		name = m[1]
	}
	return u.Host + path.Dir(u.Path) + "/" + name
}

func onPkgURL(u *url.URL) {
	log.Println("pkg url:", u.String())
	if autoMode != 1 {
		return
	}

	key := packageKey(u)
	seenMu.Lock()
	if seen[key] {
		seenMu.Unlock()
		return
	}
	seen[key] = true
	seenMu.Unlock()

	go func() {
		if err := getURLs(u.String(), "", pkgDir); err != nil {
			log.Println("link generation failed:", err)
			seenMu.Lock()
			delete(seen, key)
			seenMu.Unlock()
		}
	}()
}

func ethernetIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var fallback string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP.To4()
			if ip == nil || ip.IsLinkLocalUnicast() {
				continue
			}
			name := strings.ToLower(iface.Name)
			if strings.HasPrefix(name, "v") || strings.Contains(name, "virtual") || strings.Contains(name, "vpn") || strings.Contains(name, "docker") {
				continue
			}
			if strings.Contains(name, "ethernet") || strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") {
				return ip.String(), nil
			}
			if fallback == "" {
				fallback = ip.String()
			}
		}
	}

	if fallback != "" {
		log.Println("no ethernet interface found, using", fallback)
		return fallback, nil
	}
	return "", fmt.Errorf("no usable network interface found")
}

func isPkgRequest(r *http.Request) bool {
	return strings.HasSuffix(strings.ToLower(r.URL.Path), ".pkg")
}

func localFile(r *http.Request) (string, bool) {
	name := path.Base(r.URL.Path)
	if name == "." || name == "/" || strings.ContainsAny(name, `\/`) {
		return "", false
	}
	p := filepath.Join(pkgDir, name)
	info, err := os.Stat(p)
	if err != nil || info.IsDir() {
		return "", false
	}
	return p, true
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleHTTPS(w, r)
		return
	}

	if isPkgRequest(r) {
		if p, ok := localFile(r); ok {
			serveLocal(w, r, p)
			return
		}
		onPkgURL(r.URL)
	}

	handleHTTP(w, r)
}

func serveLocal(w http.ResponseWriter, r *http.Request, p string) {
	f, err := os.Open(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("serving local %s (range: %q)", p, r.Header.Get("Range"))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() {
		http.Error(w, "this is a proxy, not a web server", http.StatusBadRequest)
		return
	}

	r.RequestURI = ""
	r.Header.Del("Proxy-Connection")

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	dest, err := net.DialTimeout("tcp", r.Host, 15*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer dest.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}

	client, buf, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer client.Close()

	client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go func() {
		io.Copy(dest, buf)
		dest.Close()
	}()
	io.Copy(client, dest)
}

func alive(c *http.Client, u string) bool {
	resp, err := c.Head(u)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return true
		}
		if resp.StatusCode != http.StatusMethodNotAllowed {
			return false
		}
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, err = c.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent
}

func getURLs(raw, out, outDir string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid url: %s", raw)
	}

	name := path.Base(u.Path)
	dir := strings.TrimSuffix(u.Path, name)
	m := partPattern.FindStringSubmatch(name)

	var urls []string

	if m == nil {
		if !alive(probeClient, u.String()) {
			return fmt.Errorf("url is not reachable: %s", raw)
		}
		urls = append(urls, u.String())
	} else {
		prefix, ext := m[1], m[3]
		partURL := func(i int) string {
			pu := *u
			pu.Path = dir + prefix + strconv.Itoa(i) + ext
			return pu.String()
		}

		for start := 0; ; start += probeBatch {
			found := make([]bool, probeBatch)
			var wg sync.WaitGroup
			for j := 0; j < probeBatch; j++ {
				wg.Add(1)
				go func(j int) {
					defer wg.Done()
					found[j] = alive(probeClient, partURL(start+j))
				}(j)
			}
			wg.Wait()

			done := false
			for j, ok := range found {
				if !ok {
					done = true
					break
				}
				s := partURL(start + j)
				log.Println("found:", s)
				urls = append(urls, s)
			}
			if done {
				break
			}
		}
		if len(urls) == 0 {
			return fmt.Errorf("no parts found, starting from %s", dir+prefix+"0"+ext)
		}
	}

	if out == "" {
		base := name
		if m != nil {
			base = strings.TrimSuffix(m[1], "_")
		}
		out = filepath.Join(outDir, strings.TrimSuffix(base, ".pkg")+".txt")
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriter(f)
	for _, s := range urls {
		fmt.Fprintln(bw, s)
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	log.Printf("saved %d url(s) to %s", len(urls), out)
	return nil
}

func main() {
	getURL := flag.String("get-urls", "", "a .pkg url; probe all its parts (_0, _1, ...) and save them to a .txt file")
	out := flag.String("out", "", "output .txt file for --get-urls (default: <package name>.txt)")
	addr := flag.String("addr", ":8080", "proxy listen address")
	flag.StringVar(&pkgDir, "path", ".", "directory containing downloaded .pkg files")
	flag.IntVar(&autoMode, "auto", -1, "0: listen on ethernet ip:8080 and log pkg urls, 1: also generate link lists automatically")
	flag.Parse()

	if *getURL != "" {
		if err := getURLs(*getURL, *out, ""); err != nil {
			log.Fatal(err)
		}
		return
	}

	switch autoMode {
	case -1:
	case 0, 1:
		ip, err := ethernetIP()
		if err != nil {
			log.Fatal(err)
		}
		*addr = net.JoinHostPort(ip, "8080")
	default:
		log.Fatal("--auto must be 0 or 1")
	}

	abs, _ := filepath.Abs(pkgDir)
	log.Printf("proxy listening on %s, serving .pkg files from %s", *addr, abs)
	log.Fatal(http.ListenAndServe(*addr, http.HandlerFunc(handle)))
}
