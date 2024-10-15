package config

import (
	"fmt"
	"os"
	"text/template"
)

const (
	templatePath       = "./template/radvd_template.conf"
	outputPath         = "./output/"
	AdvDefaultLifetime = 540
)

type RadvdConfig struct {
	Rule                 Rule
	isDefault            bool
	FilePath             string
	AdvDefaultLifetime   int
	AdvDefaultPreference string
	Routes               []string
	Clients              []string
}

func ContainsStr(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func ContainsInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (c *Config) GenerateRadvdConfigFile() ([]RadvdConfig, error) {
	var radvdConfigs []RadvdConfig

	// policy config to radvd config
	for _, rule := range c.Rules {
		var radvdConfig RadvdConfig
		radvdConfig.Rule = rule
		radvdConfig.isDefault = false
		fileName := fmt.Sprintf("%s(%d).conf", radvdConfig.Rule.Nexthop, radvdConfig.Rule.ID)
		radvdConfig.FilePath = outputPath + fileName
		radvdConfig.AdvDefaultLifetime = AdvDefaultLifetime
		radvdConfig.AdvDefaultPreference = "medium"

		if rule.Type == "Prefixes" && !ContainsStr(rule.Prefixes, "::/0") {
			radvdConfig.Routes = append(radvdConfig.Routes, rule.Prefixes...)
		} else if len(rule.Prefixes) == 1 && rule.Prefixes[0] == "::/0" {
			radvdConfig.AdvDefaultPreference = "high"
			radvdConfig.isDefault = true
		} else {
			return nil, fmt.Errorf("invalid rule type: %s", rule.Type)
		}

		radvdConfigs = append(radvdConfigs, radvdConfig)
	}

	// apply clients
	for _, policy := range c.Policies {
		for _, ruleID := range policy.Rules {
			for i, radvdConfig := range radvdConfigs {
				if radvdConfig.Rule.ID == ruleID {
					radvdConfigs[i].Clients = append(radvdConfigs[i].Clients, policy.Clients...)
				}
			}
		}
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	// generate radvd config files
	for _, radvdConfig := range radvdConfigs {
		file, err := os.Create(radvdConfig.FilePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		err = tmpl.Execute(file, radvdConfig)
		if err != nil {
			return nil, err
		}
	}

	return radvdConfigs, nil
}
