package config

import (
	"errors"
	"os"
	"regexp"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Rule struct {
	ID          int      `yaml:"id" validate:"required"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type" validate:"oneof=FQDNs Prefixes,required"`
	FQDNs       []string `yaml:"fqdn,omitempty" validate:"dive,domain"`
	Prefixes    []string `yaml:"prefixes,omitempty" validate:"dive,cidrv6"`
	Nexthop     string   `yaml:"nexthop" validate:"ipv6,required"`
}

type Policy struct {
	ID          int      `yaml:"id" validate:"required"`
	Description string   `yaml:"description"`
	Rules       []int    `yaml:"rules" validate:"dive,rules_exist,required"`
	Clients     []string `yaml:"clients" validate:"dive,ipv6,required"`
}

type Config struct {
	Rules    []Rule   `yaml:"rules" validate:"unique=ID,required,dive" default:"[]"`
	Policies []Policy `yaml:"policies" validate:"unique=ID,required,dive" default:"[]"`
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

	for _, rule := range config.Rules {
		if err := config.validateRule(rule); err != nil {
			return nil, err
		}
	}
	for _, policy := range config.Policies {
		if err := config.validatePolicy(policy); err != nil {
			return nil, err
		}
	}

	return &config, nil
}

func LoadRadvdConfig(filePath string) ([]byte, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

type ValidationErrors = validator.ValidationErrors

var domainRegexp = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)

func (c *Config) validateRule(rule Rule) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	switch rule.Type {
	case "FQDNs":
		return errors.New("FQDNs type is not yet supported")
	case "Prefixes":
		if rule.Prefixes == nil {
			return errors.New("\"Prefixes\" type requires at least one prefix value")
		}
		if rule.FQDNs != nil {
			return errors.New("\"Prefixes\" type should not have an FQDN value")
		}
	default:
		return errors.New("invalid rule type: " + rule.Type)
	}

	validate.RegisterValidation("domain", func(fl validator.FieldLevel) bool {
		dom := fl.Field().String()
		return domainRegexp.Match([]byte(dom))
	})

	return nil
}

func (c *Config) validatePolicy(policy Policy) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterValidation("existing_rule", func(fl validator.FieldLevel) bool {
		ruleID := policy.Rules
		rules := c.Rules

		for _, rule := range rules {
			for _, ruleID := range ruleID {
				if rule.ID == ruleID {
					return true
				}
			}
		}

		return false
	})

	return nil
}
