package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/y-kzm/go-radvd-manager/internal/radvd"
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

func GetSeverList(radvd *radvd.Radvd) []string {
	severs := make(map[string]struct{})
	for _, iface := range radvd.Interfaces {
		severs[iface.Nexthop] = struct{}{}
	}

	var servers []string
	for server := range severs {
		servers = append(servers, server)
	}

	return servers
}

func (c *RadvdManagerClient) GetInstance(instance int) (*radvd.Interface, error) {
	url := c.host + "/restconf/data/radvd:interfaces/" + strconv.Itoa(instance)
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get radvd instance: %s, response: %s", resp.Status, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var iface radvd.Interface
	if err := json.Unmarshal([]byte(body), &iface); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	return &iface, nil
}

func (c *RadvdManagerClient) GetAllInstance() (*radvd.Radvd, error) {
	url := c.host + "/restconf/data/radvd:interfaces"
	resp, err := c.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get radvd instance: %s, response: %s", resp.Status, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var radvd radvd.Radvd
	if err := json.Unmarshal([]byte(body), &radvd); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	return &radvd, nil
}

func (c *RadvdManagerClient) CreateInstance(instance int, iface *radvd.Interface) error {
	jsonData, err := json.MarshalIndent(iface, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal struct to JSON: %v", err)
	}

	url := c.host + "/restconf/data/radvd:interfaces/" + strconv.Itoa(instance)
	resp, err := c.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create radvd instance: %s", resp.Status)
	}

	return nil
}

func (c *RadvdManagerClient) UpdateInstance(instance int, iface *radvd.Interface) error {
	jsonData, err := json.MarshalIndent(iface, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal struct to JSON: %v", err)
	}

	url := c.host + "/restconf/data/radvd:interfaces/" + strconv.Itoa(instance)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalf("failed to send PUT request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create radvd instance: %s", resp.Status)
	}

	return nil
}

func (c *RadvdManagerClient) DeleteInstance(instance int) error {
	url := c.host + "/restconf/data/radvd:interfaces/" + strconv.Itoa(instance)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalf("failed to send DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete radvd instance: %s", body)
	}

	return nil
}

func (c *RadvdManagerClient) DeleteAllInstance() error {
	url := c.host + "/restconf/data/radvd:interfaces"
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalf("failed to send DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete radvd instance: %s", body)
	}

	return nil
}
