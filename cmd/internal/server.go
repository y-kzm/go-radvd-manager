package internal

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"

	radvd "github.com/y-kzm/go-radvd-manager"
)

type RadvdManagerServer struct {
	http.Server
	instances []*radvd.Instance
	logger    *slog.Logger
}

func NewServer(host string, instances []*radvd.Instance, logger *slog.Logger) *RadvdManagerServer {
	if err := radvd.InitInstances(&instances); err != nil {
		logger.Error("Failed to initialize instances", "error", err.Error())
	}
	srv := &RadvdManagerServer{
		instances: instances,
		logger:    logger,
	}

	router := mux.NewRouter()
	router.HandleFunc("/rest/data/radvd:instances", srv.handleInstances).Methods("GET", "DELETE")
	router.HandleFunc("/rest/data/radvd:instances/{instance}", srv.handleInstance).Methods("GET", "POST", "PUT", "DELETE")

	srv.Addr = host
	srv.Handler = router

	return srv
}

func (s *RadvdManagerServer) handleInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.logger.Info("[GET]", "from", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		for _, i := range s.instances {
			pid, _ := radvd.GetRadvdPID(int(i.ID))
			i.PID = uint32(pid)
		}
		if err := json.NewEncoder(w).Encode(s.instances); err != nil {
			s.logger.Error("Failed to encode JSON", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	case "DELETE":
		s.logger.Info("[DELETE]", "from", r.RemoteAddr)
		for _, i := range s.instances {
			if err := radvd.StopRadvd(int(i.ID)); err != nil {
				s.logger.Error("Failed to stop radvd", "error", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			s.logger.Info("Stopped radvd", "instance", i.ID)
		}
		s.instances = []*radvd.Instance{}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		s.logger.Error("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *RadvdManagerServer) handleInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceStr := vars["instance"]
	instance, err := strconv.Atoi(instanceStr)
	if err != nil && instance <= 0 {
		s.logger.Error("Invalid Instance ID", "instance", instanceStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case "GET":
		s.logger.Info("[GET]", "from", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		for _, i := range s.instances {
			if i.ID == uint32(instance) {
				pid, _ := radvd.GetRadvdPID(int(i.ID))
				i.PID = uint32(pid)
				if err := json.NewEncoder(w).Encode(i); err != nil {
					s.logger.Error("Failed to encode JSON", "error", err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		return
	case "POST":
		s.logger.Info("[POST]", "from", r.RemoteAddr)
		// Check if the instance already exists
		for _, i := range s.instances {
			if i.ID == uint32(instance) {
				s.logger.Error("Instance already exists", "instance", instance)
				w.WriteHeader(http.StatusConflict)
				return
			}
		}
		var new radvd.Instance
		if err := json.NewDecoder(r.Body).Decode(&new); err != nil {
			s.logger.Error("Failed to decode JSON", "error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if new.ID != uint32(instance) {
			s.logger.Error("Instance ID mismatch", "instance", instance, "instance in body", new.ID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// generate radvd config file
		if err := radvd.GenerateRadvdConfigFile(&new, radvd.RadvdConfPath); err != nil {
			s.logger.Error("Failed to generate radvd config file", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := radvd.CheckRadvdConfig(int(new.ID)); err != nil {
			s.logger.Error("Failed to check radvd config", "error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// start radvd process
		if err := radvd.StartRadvd(int(new.ID)); err != nil {
			s.logger.Error("Failed to start radvd", "error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.instances = append(s.instances, &new)
		w.WriteHeader(http.StatusCreated)
		return
	case "PUT":
	case "DELETE":
	default:
		s.logger.Error("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *RadvdManagerServer) CleanUp() error {
	for _, i := range s.instances {
		if err := radvd.StopRadvd(int(i.ID)); err != nil {
			s.logger.Error("Failed to stop radvd", "error", err.Error())
			continue
		}
		s.logger.Info("Stopped radvd", "instance", i.ID)
	}
	s.instances = []*radvd.Instance{}

	files, err := filepath.Glob("/etc/radvd.d/*")
	if err != nil {
		s.logger.Error("Failed to glob files in /etc/radvd.d/", "error", err.Error())
	} else {
		for _, file := range files {
			if err := os.Remove(file); err != nil {
				s.logger.Error("Failed to remove config file", "file", file, "error", err.Error())
			}
		}
	}
	pidFiles, err := filepath.Glob("/var/run/radvd/radvd.*")
	if err != nil {
		s.logger.Error("Failed to glob PID files in /var/run/radvd/", "error", err.Error())
	} else {
		for _, pidFile := range pidFiles {
			if err := os.Remove(pidFile); err != nil {
				s.logger.Error("Failed to remove PID file", "file", pidFile, "error", err.Error())
			}
		}
	}

	return nil
}
