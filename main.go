package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type RadvdInfo struct {
	ID     string `json:"id"`
	Config string `json:"config"`
}

var (
	radvdInstances = make(map[string]*RadvdInfo)
	mutex          = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/daemons", handleRadvdMgmt)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handle REST API
func handleRadvdMgmt(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		log.Printf("GET request from %s", r.RemoteAddr)
		handleGetRadvd(w, r)
	case "POST":
		log.Printf("POST request from %s", r.RemoteAddr)
		handleCreateRadvd(w, r)
	case "PUT":
		log.Printf("PUT request from %s", r.RemoteAddr)
		handleCreateRadvd(w, r)
	case "DELETE":
		log.Printf("DELETE request from %s", r.RemoteAddr)
		handleDeleteRadvd(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Delete radvd instance
func handleDeleteRadvd(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := radvdInstances[id]; !exists {
		http.Error(w, "Instance not found", http.StatusNotFound)
		return
	}

	delete(radvdInstances, id)

	pidFile := fmt.Sprintf("/var/run/radvd_%s.pid", id)
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, "Failed to read PID file", http.StatusInternalServerError)
			return
		}
	} else {
		pid := strings.TrimSpace(string(pidData))
		if err := exec.Command("kill", pid).Run(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to kill radvd process: %v", err), http.StatusInternalServerError)
			return
		}
		if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Failed to remove PID file: %v", err), http.StatusInternalServerError)
			return
		}
	}

	configPath := fmt.Sprintf("/etc/radvd/radvd_%s.conf", id)
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Failed to remove config file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("radvd instance %s deleted", id)))
}

// Create radvd instance
func handleCreateRadvd(w http.ResponseWriter, r *http.Request) {
	var radvdInfo RadvdInfo

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &radvdInfo); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := radvdInstances[radvdInfo.ID]; exists {
		http.Error(w, "Instance already exists", http.StatusConflict)
		return
	}

	radvdInstances[radvdInfo.ID] = &radvdInfo

	// Create the config file for radvd
	configPath := fmt.Sprintf("/etc/radvd/radvd_%s.conf", radvdInfo.ID)
	if err := os.WriteFile(configPath, []byte(radvdInfo.Config), 0644); err != nil {
		http.Error(w, "Failed to write config file", http.StatusInternalServerError)
		return
	}

	// Start the radvd process
	if err := startRadvd(radvdInfo.ID, configPath); err != nil {
		http.Error(w, "Failed to start radvd", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("radvd instance %s created", radvdInfo.ID)))
}

// Get radvd instance configuration
func handleGetRadvd(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		jsonData, _ := json.Marshal(radvdInstances)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}

	mutex.Lock()
	defer mutex.Unlock()

	if config, exists := radvdInstances[id]; exists {
		jsonData, _ := json.Marshal(config)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		http.Error(w, "Instance not found", http.StatusNotFound)
	}
}

// Start or restart radvd instance
func startRadvd(id string, configPath string) error {
	pidFile := fmt.Sprintf("/var/run/radvd_%s.pid", id)
	cmd := exec.Command("radvd", "-C", configPath, "-p", pidFile)
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Printf("radvd instance %s started with PID %d", id, cmd.Process.Pid)
	return cmd.Wait()
}
