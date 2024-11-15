/**
 * Parse radvd configuration files.
 * Convert radvd configuration files to radvd.Interface struct.
 *
 */
package config

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/y-kzm/go-radvd-manager/internal/radvd"
)

func parseBool(value string) bool {
	return value == "yes" || value == "on" || value == "true"
}

func parseUint32(value string) uint32 {
	num, _ := strconv.ParseUint(value, 10, 32)
	return uint32(num)
}

func parseRadvdConf(filePath string, instance int) (*radvd.Interface, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var iface radvd.Interface
	var currentPrefix radvd.Prefix
	var currentRdnss radvd.RDNSS
	var currentRoute radvd.Route
	iface.Instance = uint32(instance)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(line, "interface"):
			fields := strings.Fields(line)
			iface.Name = fields[1]

		case strings.HasPrefix(line, "AdvSendAdvert"):
			iface.AdvSendAdvert = parseBool(strings.TrimSuffix(line[len("AdvSendAdvert "):], ";"))

		case strings.HasPrefix(line, "MinRtrAdvInterval"):
			iface.MinRtrAdvInterval = parseUint32(strings.TrimSuffix(line[len("MinRtrAdvInterval "):], ";"))

		case strings.HasPrefix(line, "MaxRtrAdvInterval"):
			iface.MaxRtrAdvInterval = parseUint32(strings.TrimSuffix(line[len("MaxRtrAdvInterval "):], ";"))

		case strings.HasPrefix(line, "AdvManagedFlag"):
			iface.AdvManagedFlag = parseBool(strings.TrimSuffix(line[len("AdvManagedFlag "):], ";"))

		case strings.HasPrefix(line, "AdvOtherConfigFlag"):
			iface.AdvOtherConfigFlag = parseBool(strings.TrimSuffix(line[len("AdvOtherConfigFlag "):], ";"))

		case strings.HasPrefix(line, "AdvDefaultLifetime"):
			iface.AdvDefaultLifetime = parseUint32(strings.TrimSuffix(line[len("AdvDefaultLifetime "):], ";"))

		case strings.HasPrefix(line, "AdvDefaultPreference"):
			iface.AdvDefaultPreference = strings.TrimSuffix(line[len("AdvDefaultPreference "):], ";")

		case strings.HasPrefix(line, "prefix"):
			fields := strings.Fields(line)
			currentPrefix = radvd.Prefix{Prefix: fields[1]}

		case strings.HasPrefix(line, "AdvOnLink"):
			currentPrefix.AdvOnLink = parseBool(strings.TrimSuffix(line[len("AdvOnLink "):], ";"))

		case strings.HasPrefix(line, "AdvAutonomous"):
			currentPrefix.AdvAutonomous = parseBool(strings.TrimSuffix(line[len("AdvAutonomous "):], ";"))

		case strings.HasPrefix(line, "AdvRouterAddr"):
			currentPrefix.AdvRouterAddr = parseBool(strings.TrimSuffix(line[len("AdvRouterAddr "):], ";"))

		case strings.HasPrefix(line, "AdvValidLifetime"):
			currentPrefix.AdvValidLifetime = parseUint32(strings.TrimSuffix(line[len("AdvValidLifetime "):], ";"))

		case line == "};":
			if currentPrefix.Prefix != "" {
				iface.Prefixes = append(iface.Prefixes, currentPrefix)
				currentPrefix = radvd.Prefix{}
			}
			if currentRdnss.Address != "" {
				iface.Rdnss = append(iface.Rdnss, currentRdnss)
				currentRdnss = radvd.RDNSS{}
			}
			if currentRoute.Route != "" {
				iface.Routes = append(iface.Routes, currentRoute)
				currentRoute = radvd.Route{}
			}

		case strings.HasPrefix(line, "RDNSS"):
			fields := strings.Fields(line)
			currentRdnss = radvd.RDNSS{Address: fields[1]}

		case strings.HasPrefix(line, "AdvRDNSSLifetime"):
			currentRdnss.AdvRdnssLifetime = parseUint32(strings.TrimSuffix(line[len("AdvRDNSSLifetime "):], ";"))

		case strings.HasPrefix(line, "route"):
			fields := strings.Fields(line)
			currentRoute = radvd.Route{Route: fields[1]}

		case strings.HasPrefix(line, "AdvRouteLifetime"):
			currentRoute.AdvRouteLifetime = parseUint32(strings.TrimSuffix(line[len("AdvRouteLifetime "):], ";"))

		case strings.HasPrefix(line, "AdvRoutePreference"):
			currentRoute.AdvRoutePreference = strings.TrimSuffix(line[len("AdvRoutePreference "):], ";")

		case strings.HasPrefix(line, "clients"):
			for scanner.Scan() {
				clientLine := strings.TrimSpace(scanner.Text())
				if clientLine == "};" {
					break
				}
				iface.Clients = append(iface.Clients, strings.TrimSuffix(clientLine, ";"))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %v", err)
	}

	return &iface, nil
}

// TODO: Check correspondence with /var/run/radvd/*.
func ParseRadvdConfigs() (*radvd.Radvd, error) {
	var radvd radvd.Radvd

	iface, err := parseRadvdConf("/etc/radvd.conf", 0)
	if err != nil {
		return nil, err
	}
	radvd.Interfaces = append(radvd.Interfaces, iface)

	dirPath := "/etc/radvd.d"
	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".conf") {
			base := strings.TrimSuffix(d.Name(), ".conf")
			instance, err := strconv.Atoi(base)
			if err != nil {
				fmt.Printf("Failed to convert instance number from file name: %v\n", err)
				return nil
			}
			iface, err := parseRadvdConf(path, instance)
			if err != nil {
				return err
			}
			radvd.Interfaces = append(radvd.Interfaces, iface)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &radvd, nil
}
