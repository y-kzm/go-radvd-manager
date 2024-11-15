package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"

	"github.com/y-kzm/go-radvd-manager/internal/radvd"
)

type Policy struct {
	ID          int      `yaml:"id" validate:"required"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type" validate:"oneof=FQDNs Prefixes,required"`
	FQDNs       []string `yaml:"fqdn,omitempty" validate:"dive,domain"`
	Prefixes    []string `yaml:"prefixes,omitempty" validate:"dive,cidrv6"`
	Nexthop     string   `yaml:"nexthop" validate:"ipv6,required"`
}

type Group struct {
	ID          int      `yaml:"id" validate:"required"`
	Description string   `yaml:"description"`
	Policies    []int    `yaml:"policies" validate:"dive,rules_exist,required"`
	Members     []string `yaml:"members" validate:"dive,ipv6,required"`
}

type Config struct {
	Policies []Policy `yaml:"policies" validate:"unique=ID,required,dive" default:"[]"`
	Groups   []Group  `yaml:"groups" validate:"unique=ID,required,dive" default:"[]"`
}

type radvdInterfaceAlias = radvd.Interface
type radvdPrefixAlias = radvd.Prefix
type radvdRDNSSAlias = radvd.RDNSS
type radvdRouteAlias = radvd.Route

func containsStr(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func ConfigToRadvd(config *Config) (*radvd.Radvd, error) {
	var param *radvd.Radvd
	var radvd radvd.Radvd

	param, err := LoadDefaultParameterFile()
	if err != nil {
		log.Fatalf("Failed to marshal radvd to JSON: %v", err)
	}

	for _, policy := range config.Policies {
		var iface radvdInterfaceAlias
		for _, defaultParam := range param.Interfaces {
			if defaultParam.Nexthop == policy.Nexthop {
				iface = *defaultParam
				break
			}
		}
		iface.Instance = uint32(policy.ID)

		if policy.Type == "Prefixes" && !containsStr(policy.Prefixes, "::/0") {
			for _, prefix := range policy.Prefixes {
				route := radvdRouteAlias{
					Route:              prefix,
					AdvRouteLifetime:   1800,
					AdvRoutePreference: "medium",
				}
				iface.Routes = append(iface.Routes, route)
			}
		} else if len(policy.Prefixes) == 1 && policy.Prefixes[0] == "::/0" {
			// "::/0" must be the only element in the list
			iface.AdvDefaultPreference = "high"
		} else {
			return nil, fmt.Errorf("invalid policy type: %s", policy.Type)
		}
		radvd.Interfaces = append(radvd.Interfaces, &iface)
	}

	for _, group := range config.Groups {
		for _, policy := range group.Policies {
			for _, iface := range radvd.Interfaces {
				if iface.Instance == uint32(policy) {
					iface.Clients = append(iface.Clients, group.Members...)
				}
			}
		}
	}

	for _, iface := range radvd.Interfaces {
		err := GenerateRadvdConfigFile2(iface, "./output/")
		if err != nil {
			return nil, err
		}
	}

	return &radvd, nil
}

func LoadPolicyConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	for _, policy := range config.Policies {
		if err := config.validatePolicy(policy); err != nil {
			return nil, err
		}
	}
	for _, group := range config.Groups {
		if err := config.validateGroup(group); err != nil {
			return nil, err
		}
	}

	return &config, nil
}

type ValidationErrors = validator.ValidationErrors

var domainRegexp = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)

func (c *Config) validatePolicy(policy Policy) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	switch policy.Type {
	case "FQDNs":
		return errors.New("FQDNs type is not yet supported")
	case "Prefixes":
		if policy.Prefixes == nil {
			return errors.New("\"Prefixes\" type requires at least one prefix value")
		}
		if policy.FQDNs != nil {
			return errors.New("\"Prefixes\" type should not have an FQDN value")
		}
	default:
		return errors.New("invalid rule type: " + policy.Type)
	}

	validate.RegisterValidation("domain", func(fl validator.FieldLevel) bool {
		dom := fl.Field().String()
		return domainRegexp.Match([]byte(dom))
	})

	return nil
}

func (c *Config) validateGroup(group Group) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterValidation("existing_rule", func(fl validator.FieldLevel) bool {

		for _, policy := range c.Policies {
			for _, applyPolicy := range group.Policies {
				if policy.ID == applyPolicy {
					return true
				}
			}
		}

		return false
	})

	return nil
}
