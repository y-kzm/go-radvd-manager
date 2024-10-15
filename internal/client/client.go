package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/y-kzm/go-radvd-manager/internal/config"
)

type RadvdManagerClient struct {
	*http.Client
	host   string
	Server string
	Port   int
}

func NewClient(host string, server string, port int) *RadvdManagerClient {
	return &RadvdManagerClient{
		Client: &http.Client{},
		host:   host,
		Server: server,
		Port:   port,
	}
}

func (c *RadvdManagerClient) Create(id int, config string) error {
	data := map[string]string{
		"config": config,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := c.host + "/daemons/" + strconv.Itoa(id)
	resp, err := c.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create radvd instance: %s", body)
	}

	return nil
}

func (c *RadvdManagerClient) Get(id int) (string, error) {
	url := c.host + "/daemons/" + strconv.Itoa(id)
	resp, err := c.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get radvd instance: %s, response: %s", resp.Status, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *RadvdManagerClient) Delete(id int) error {
	req, err := http.NewRequest(http.MethodDelete, c.host+"/daemons/"+strconv.Itoa(id), nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete radvd instance: %s", body)
	}

	return nil
}

func GetSeverList(radvdConfigs []config.RadvdConfig) []string {
	uniqueSevers := make(map[string]struct{})
	for _, radvdConfig := range radvdConfigs {
		uniqueSevers[radvdConfig.Rule.Nexthop] = struct{}{}
	}

	var servers []string
	for server := range uniqueSevers {
		servers = append(servers, server)
	}

	return servers
}
