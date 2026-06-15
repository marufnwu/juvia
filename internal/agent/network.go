package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const publicIPTimeout = 10 * time.Second

var ipDiscoveryURLs = []string{
	"https://ifconfig.me/ip",
	"https://api.ipify.org",
	"https://checkip.amazonaws.com",
}

// HandleNetworkGetPublicIP discovers the server's public IPv4 address.
// It tries multiple external services and falls back to environment variables
// or local network interfaces.
func HandleNetworkGetPublicIP(ctx context.Context, params json.RawMessage) (interface{}, error) {
	ip := discoverPublicIP(ctx)
	if ip == "" {
		return nil, fmt.Errorf("unable to determine public IP address")
	}
	return map[string]interface{}{
		"ip": ip,
	}, nil
}

func discoverPublicIP(ctx context.Context) string {
	if envIP := strings.TrimSpace(os.Getenv("JUVIA_PUBLIC_IP")); envIP != "" {
		return envIP
	}

	for _, url := range ipDiscoveryURLs {
		ip := fetchIP(ctx, url)
		if ip != "" {
			return ip
		}
	}

	return ""
}

func fetchIP(ctx context.Context, url string) string {
	client := &http.Client{Timeout: publicIPTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return ""
	}

	ip := strings.TrimSpace(string(body))
	if ip == "" || strings.Contains(ip, "error") {
		return ""
	}
	return ip
}
