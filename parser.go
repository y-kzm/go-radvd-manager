package radvd_manager

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultRadvdInstanceID = 0
)

func parseRadvdConf(filePath string, id int) (*Instance, error) {
	var instance Instance
	var prefix Prefix
	var rdnss RDNSS
	var route Route

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()
	instance.ID = uint32(id)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(line, "interface"):
			fields := strings.Fields(line)
			instance.Name = fields[1]

		case strings.HasPrefix(line, "AdvSendAdvert"):
			instance.AdvSendAdvert = parseBool(strings.TrimSuffix(line[len("AdvSendAdvert "):], ";"))

		case strings.HasPrefix(line, "MinRtrAdvInterval"):
			instance.MinRtrAdvInterval = parseUint32(strings.TrimSuffix(line[len("MinRtrAdvInterval "):], ";"))

		case strings.HasPrefix(line, "MaxRtrAdvInterval"):
			instance.MaxRtrAdvInterval = parseUint32(strings.TrimSuffix(line[len("MaxRtrAdvInterval "):], ";"))

		case strings.HasPrefix(line, "AdvManagedFlag"):
			instance.AdvManagedFlag = parseBool(strings.TrimSuffix(line[len("AdvManagedFlag "):], ";"))

		case strings.HasPrefix(line, "AdvOtherConfigFlag"):
			instance.AdvOtherConfigFlag = parseBool(strings.TrimSuffix(line[len("AdvOtherConfigFlag "):], ";"))

		case strings.HasPrefix(line, "AdvDefaultLifetime"):
			instance.AdvDefaultLifetime = parseUint32(strings.TrimSuffix(line[len("AdvDefaultLifetime "):], ";"))

		case strings.HasPrefix(line, "AdvDefaultPreference"):
			instance.AdvDefaultPreference = strings.TrimSuffix(line[len("AdvDefaultPreference "):], ";")

		case strings.HasPrefix(line, "prefix"):
			fields := strings.Fields(line)
			prefix = Prefix{Prefix: fields[1]}

		case strings.HasPrefix(line, "AdvOnLink"):
			prefix.AdvOnLink = parseBool(strings.TrimSuffix(line[len("AdvOnLink "):], ";"))

		case strings.HasPrefix(line, "AdvAutonomous"):
			prefix.AdvAutonomous = parseBool(strings.TrimSuffix(line[len("AdvAutonomous "):], ";"))

		case strings.HasPrefix(line, "AdvRouterAddr"):
			prefix.AdvRouterAddr = parseBool(strings.TrimSuffix(line[len("AdvRouterAddr "):], ";"))

		case strings.HasPrefix(line, "AdvValidLifetime"):
			prefix.AdvValidLifetime = parseUint32(strings.TrimSuffix(line[len("AdvValidLifetime "):], ";"))

		case line == "};":
			if prefix.Prefix != "" {
				instance.Prefixes = append(instance.Prefixes, prefix)
				prefix = Prefix{}
			}
			if rdnss.Address != "" {
				instance.Rdnss = append(instance.Rdnss, rdnss)
				rdnss = RDNSS{}
			}
			if route.Route != "" {
				instance.Routes = append(instance.Routes, route)
				route = Route{}
			}

		case strings.HasPrefix(line, "RDNSS"):
			fields := strings.Fields(line)
			rdnss = RDNSS{Address: fields[1]}

		case strings.HasPrefix(line, "AdvRDNSSLifetime"):
			rdnss.AdvRdnssLifetime = parseUint32(strings.TrimSuffix(line[len("AdvRDNSSLifetime "):], ";"))

		case strings.HasPrefix(line, "route"):
			fields := strings.Fields(line)
			route = Route{Route: fields[1]}

		case strings.HasPrefix(line, "AdvRouteLifetime"):
			route.AdvRouteLifetime = parseUint32(strings.TrimSuffix(line[len("AdvRouteLifetime "):], ";"))

		case strings.HasPrefix(line, "AdvRoutePreference"):
			route.AdvRoutePreference = strings.TrimSuffix(line[len("AdvRoutePreference "):], ";")

		case strings.HasPrefix(line, "clients"):
			for scanner.Scan() {
				clientLine := strings.TrimSpace(scanner.Text())
				if clientLine == "};" {
					break
				}
				instance.Clients = append(instance.Clients, strings.TrimSuffix(clientLine, ";"))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %v", err)
	}

	return &instance, nil
}

func InitInstances(instances *[]*Instance) error {
	defaultInstance, err := parseRadvdConf(defaultRadvdCondFile, defaultRadvdInstanceID)
	if err != nil {
		return err
	}
	*instances = append(*instances, defaultInstance)
	err = filepath.WalkDir(RadvdConfPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".conf") {
			base := strings.TrimSuffix(d.Name(), ".conf")
			id, err := strconv.Atoi(base)
			if err != nil {
				fmt.Printf("Failed to convert instance number from file name: %v\n", err)
				return nil
			}
			instance, err := parseRadvdConf(path, id)
			if err != nil {
				return err
			}
			*instances = append(*instances, instance)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func parseBool(value string) bool {
	return value == "yes" || value == "on" || value == "true"
}

func parseUint32(value string) uint32 {
	num, _ := strconv.ParseUint(value, 10, 32)
	return uint32(num)
}
