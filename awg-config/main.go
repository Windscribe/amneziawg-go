package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	sockPath = "/var/run/amneziawg"
)

func main() {
	var (
		interfaceName = flag.String("i", "utun0", "Interface name")
		configFile    = flag.String("c", "", "Configuration file path")
		showConfig    = flag.Bool("show", false, "Show current configuration")
	)
	flag.Parse()

	socketPath := filepath.Join(sockPath, *interfaceName+".sock")

	if *showConfig {
		if err := getConfig(socketPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *configFile == "" {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintf(os.Stderr, "  Set config: %s -i <interface> -c <config-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Show config: %s -i <interface> -show\n", os.Args[0])
		os.Exit(1)
	}

	if err := setConfig(socketPath, *configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration applied successfully")
}

func setConfig(socketPath, configFile string) error {
	// Read and parse config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Convert WireGuard INI format to UAPI format
	uapiConfig, err := parseConfig(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Debug: print the UAPI config
	fmt.Println("=== UAPI Config ===")
	fmt.Print(uapiConfig)
	fmt.Println("=== End Config ===")

	// Connect to UAPI socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to socket %s: %w", socketPath, err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Send set operation
	if _, err := writer.WriteString("set=1\n"); err != nil {
		return fmt.Errorf("failed to write set command: %w", err)
	}

	// Send config data
	if _, err := writer.WriteString(uapiConfig); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Send empty line to terminate
	if _, err := writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write terminator: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	// Read response
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(response)
	if response != "errno=0" {
		// Read additional line
		reader.ReadString('\n')
		return fmt.Errorf("configuration failed: %s", response)
	}

	return nil
}

func parseConfig(data string) (string, error) {
	var result strings.Builder
	var inPeerSection bool

	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle sections
		if strings.HasPrefix(line, "[") {
			if strings.Contains(line, "[Peer]") {
				inPeerSection = true
			} else if strings.Contains(line, "[Interface]") {
				inPeerSection = false
			}
			continue
		}

		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Convert to UAPI format
		uapiKey := convertKey(key, inPeerSection)
		if uapiKey == "" {
			continue
		}

		// Convert base64 keys to hex
		if uapiKey == "private_key" || uapiKey == "public_key" || uapiKey == "preshared_key" {
			hexValue, err := base64ToHex(value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to convert %s: %v\n", uapiKey, err)
				continue
			}
			value = hexValue
		}

		// Handle special cases
		if uapiKey == "public_key" && inPeerSection {
			result.WriteString(fmt.Sprintf("%s=%s\n", uapiKey, value))
		} else if uapiKey == "allowed_ip" {
			// Split multiple IPs if comma-separated
			for _, ip := range strings.Split(value, ",") {
				ip = strings.TrimSpace(ip)
				if ip != "" {
					result.WriteString(fmt.Sprintf("%s=%s\n", uapiKey, ip))
				}
			}
		} else {
			result.WriteString(fmt.Sprintf("%s=%s\n", uapiKey, value))
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}

func base64ToHex(b64 string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(decoded), nil
}

func convertKey(key string, inPeerSection bool) string {
	key = strings.ToLower(key)

	// Map WireGuard config keys to UAPI keys
	keyMap := map[string]string{
		"privatekey":          "private_key",
		"publickey":           "public_key",
		"presharedkey":        "preshared_key",
		"listenport":          "listen_port",
		"address":             "", // Ignored - set via ip command
		"dns":                 "", // Ignored - set separately
		"mtu":                 "", // Ignored - set via ip command
		"endpoint":            "endpoint",
		"allowedips":          "allowed_ip",
		"persistentkeepalive": "persistent_keepalive_interval",
		"jc":                  "jc",
		"jmin":                "jmin",
		"jmax":                "jmax",
		"s1":                  "s1",
		"s2":                  "s2",
		"s3":                  "s3",
		"s4":                  "s4",
		"h1":                  "h1",
		"h2":                  "h2",
		"h3":                  "h3",
		"h4":                  "h4",
		"i1":                  "i1",
		"i2":                  "i2",
		"i3":                  "i3",
		"i4":                  "i4",
		"i5":                  "i5",
	}

	if uapiKey, ok := keyMap[key]; ok {
		return uapiKey
	}

	return ""
}

func getConfig(socketPath string) error {
	// Connect to UAPI socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to socket %s: %w", socketPath, err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Send get operation
	if _, err := writer.WriteString("get=1\n\n"); err != nil {
		return fmt.Errorf("failed to write get command: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	// Read configuration
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "errno=") {
			break
		}
		fmt.Println(line)
	}

	return nil
}
