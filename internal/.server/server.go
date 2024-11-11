package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
)

type RadvdInstance struct {
	ID     int
	Config string `json:"config"`
}

var (
	radvdInstances = make(map[int]*RadvdInstance)
	mutex          = &sync.Mutex{}
)

type RadvdManagerServer struct {
	http.Server
	logger *slog.Logger
}

func NewServer(host string, logger *slog.Logger) *RadvdManagerServer {
	srv := &RadvdManagerServer{
		logger: logger,
	}

	router := mux.NewRouter()
	router.HandleFunc("/daemons", srv.handleRadvdInstances).Methods("GET")
	router.HandleFunc("/daemons/{id}", srv.handleRadvdInstanceId).Methods("GET", "POST", "PUT", "DELETE")

	srv.Addr = host
	srv.Handler = router

	return srv
}

func (s *RadvdManagerServer) handleRadvdInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.logger.Info("GET request", "remote_addr", r.RemoteAddr)
		s.getRadvdInstance(w, r, nil)
	default:
		s.logger.Error("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *RadvdManagerServer) handleRadvdInstanceId(w http.ResponseWriter, r *http.Request) {
	//s.logger.Info("Request", "method", r.Method, "url", r.URL.Path)
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		s.logger.Error("ID not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		s.logger.Error("Invalid ID", "id", idStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.logger.Info("GET request", "remote_addr", r.RemoteAddr)
		s.getRadvdInstance(w, r, &id)
	case "POST":
		s.logger.Info("POST request", "remote_addr", r.RemoteAddr)
		s.createRadvdInstance(w, r, &id)
	case "PUT":
		s.logger.Info("PUT request", "remote_addr", r.RemoteAddr)
	case "DELETE":
		s.logger.Info("DELETE request", "remote_addr", r.RemoteAddr)
		s.deleteRadvdInstance(w, r, &id)
	default:
		s.logger.Error("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *RadvdManagerServer) getRadvdInstance(w http.ResponseWriter, r *http.Request, id *int) {
	defer r.Body.Close()

	var data map[string]string

	if id != nil {
		config, err := os.ReadFile("/etc/radvd.d/" + strconv.Itoa(*id) + ".conf")
		if err != nil {
			s.logger.Error("Failed to read config file", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data = map[string]string{
			"config": string(config),
		}
	} else {
		return
	}

	j, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to marshal JSON", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func (s *RadvdManagerServer) createRadvdInstance(w http.ResponseWriter, r *http.Request, id *int) {
	defer r.Body.Close()

	var radvd RadvdInstance

	if err := json.NewDecoder(r.Body).Decode(&radvd); err != nil {
		s.logger.Error("Failed to decode JSON", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if id != nil {
		radvd.ID = *id
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := radvdInstances[radvd.ID]; exists {
		s.logger.Error("Instance already exists", "id", radvd.ID)
		w.WriteHeader(http.StatusConflict)
		return
	}

	radvdInstances[radvd.ID] = &radvd

	// Create the config file for radvd
	configPath := "/etc/radvd.d/" + strconv.Itoa(radvd.ID) + ".conf"
	if err := os.WriteFile(configPath, []byte(radvd.Config), 0644); err != nil {
		s.logger.Error("Failed to write config file", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.startRadvd(radvd.ID, configPath); err != nil {
		s.logger.Error("Failed to start radvd", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("radvd instance " + strconv.Itoa(radvd.ID) + " created"))
}

func (s *RadvdManagerServer) deleteRadvdInstance(w http.ResponseWriter, r *http.Request, id *int) {
	defer r.Body.Close()

	if id == nil {
		s.logger.Error("ID not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if err := s.stopRadvd(*id); err != nil {
		s.logger.Error("Failed to stop radvd", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete the config file for radvd
	configPath := "/etc/radvd.d/" + strconv.Itoa(*id) + ".conf"
	if err := os.Remove(configPath); err != nil {
		s.logger.Error("Failed to remove config file", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	delete(radvdInstances, *id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("radvd instance " + strconv.Itoa(*id) + " deleted"))
}

func (s *RadvdManagerServer) startRadvd(id int, configPath string) error {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(id) + ".pid"
	cmd := exec.Command("radvd", "-C", configPath, "-p", pidFile)
	if err := cmd.Start(); err != nil {
		s.logger.Error("Failed to start radvd", "error", err.Error())
		return err
	}

	s.logger.Info("radvd started", "id", id, "pid", cmd.Process.Pid)
	return nil
}

func (s *RadvdManagerServer) stopRadvd(id int) error {
	pidFile := "/var/run/radvd/radvd." + strconv.Itoa(id) + ".pid"
	file, err := os.Open(pidFile)
	if err != nil {
		fmt.Println("Error opening PID file:", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		pid, err := strconv.Atoi(scanner.Text())
		if err != nil {
			s.logger.Error("Error converting PID", "error", err.Error())
			return err
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			s.logger.Error("Error finding radvd process", "error", err.Error())
			return err
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			s.logger.Error("Error stopping radvd", "error", err.Error())
		} else {
			s.logger.Info("radvd stopped", "id", id, "pid", pid)
		}

		s.logger.Info("radvd stopped", "id", id, "pid", pid)
		return nil
	} else if err := scanner.Err(); err != nil {
		s.logger.Error("Error reading PID file", "error", err.Error())
		return err
	}

	return nil
}
