package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	radvd "github.com/y-kzm/go-radvd-manager"
)

const (
	pathInstance  = "/rest/data/radvd:instances/"
	pathInstances = "/rest/data/radvd:instances"
)

type RadvdManagerClient struct {
	*http.Client
	host            string
	Server          string
	Port            int
	RemoteInstances []*radvd.Instance
}

func NewClient(host string, server string, port int) *RadvdManagerClient {
	return &RadvdManagerClient{
		Client:          &http.Client{},
		host:            host,
		Server:          server,
		Port:            port,
		RemoteInstances: []*radvd.Instance{},
	}
}

func GetSiteExitRouters(instances []*radvd.Instance) []string {
	routers := make(map[string]struct{})
	for _, i := range instances {
		routers[i.RouterID] = struct{}{}
	}
	// Convert map to slice
	var uniqueRouters []string
	for i := range routers {
		uniqueRouters = append(uniqueRouters, i)
	}

	return uniqueRouters
}

// [GET] /rest/data/radvd:instances/{instance}
func (c *RadvdManagerClient) GetInstance(id int) (*radvd.Instance, error) {
	url := c.host + pathInstance + strconv.Itoa(id)
	res, err := c.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to get radvd instance: %s, response: %s", res.Status, body)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var instance radvd.Instance
	if err := json.Unmarshal([]byte(body), &instance); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	return &instance, nil
}

// [GET] /rest/data/radvd:instances
func (c *RadvdManagerClient) GetInstances() error {
	url := c.host + pathInstances
	res, err := c.Client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to get radvd instance: %s, response: %s", res.Status, body)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(body), &c.RemoteInstances); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	return nil
}

// [POST] /rest/data/radvd:instances/{instance}
func (c *RadvdManagerClient) CreateInstance(id int, instance *radvd.Instance) error {
	jsonData, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal struct to JSON: %v", err)
	}
	url := c.host + pathInstance + strconv.Itoa(id)
	res, err := c.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create radvd instance: %s", res.Status)
	}

	return nil
}

// [PUT] /rest/data/radvd:instances/{instance}
func (c *RadvdManagerClient) UpdateInstance(id int, iface *radvd.Instance) error {

	return nil
}

// [DELETE] /rest/data/radvd:instances/{instance}
func (c *RadvdManagerClient) DeleteInstance(instance int) error {

	return nil
}

// [DELETE] /rest/data/radvd:instances
func (c *RadvdManagerClient) DeleteInstances() error {
	url := c.host + pathInstances
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
