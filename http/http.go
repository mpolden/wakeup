package http

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/mpolden/wakeup/wol"
)

type wakeFunc func(net.IP, net.HardwareAddr) error

type Server struct {
	SourceIP  net.IP
	StaticDir string
	cacheFile string
	mu        sync.RWMutex
	wakeFunc
}

type Error struct {
	err     error
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type Devices struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	Name       string `json:"name,omitempty"`
	MACAddress string `json:"macAddress"`
}

func (d *Devices) add(device Device) {
	for _, v := range d.Devices {
		if device.MACAddress == v.MACAddress {
			return
		}
	}
	d.Devices = append(d.Devices, device)
}

func (d *Devices) remove(device Device) {
	var keep []Device
	for _, v := range d.Devices {
		if device.MACAddress == v.MACAddress {
			continue
		}
		keep = append(keep, v)
	}
	d.Devices = keep
}

func New(cacheFile string) *Server { return &Server{cacheFile: cacheFile, wakeFunc: wol.Wake} }

func (s *Server) readDevices() (*Devices, error) {
	f, err := os.OpenFile(s.cacheFile, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var i Devices
	if len(data) == 0 {
		i.Devices = make([]Device, 0)
		return &i, nil
	}
	if err := json.Unmarshal(data, &i); err != nil {
		return nil, err
	}
	if i.Devices == nil {
		i.Devices = make([]Device, 0)
	}
	sort.Slice(i.Devices, func(j, k int) bool { return i.Devices[j].MACAddress < i.Devices[k].MACAddress })
	return &i, nil
}

func (s *Server) writeDevice(device Device, add bool) error {
	i, err := s.readDevices()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.cacheFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if add {
		i.add(device)
	} else {
		i.remove(device)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(i); err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (s *Server) defaultHandler(w http.ResponseWriter, r *http.Request) (interface{}, *Error) {
	defer r.Body.Close()
	if r.Method == http.MethodGet {
		s.mu.RLock()
		defer s.mu.RUnlock()
		i, err := s.readDevices()
		if err != nil {
			return nil, &Error{err: err, Status: http.StatusInternalServerError, Message: "Could not unmarshal JSON"}
		}
		return i, nil
	}
	add := r.Method == http.MethodPost
	remove := r.Method == http.MethodDelete
	if add || remove {
		dec := json.NewDecoder(r.Body)
		var device Device
		if err := dec.Decode(&device); err != nil {
			return nil, &Error{Status: http.StatusBadRequest, Message: "Malformed JSON"}
		}
		if add {
			macAddress, err := net.ParseMAC(device.MACAddress)
			if err != nil {
				return nil, &Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid MAC address: %s", device.MACAddress)}
			}
			if err := s.wakeFunc(s.SourceIP, macAddress); err != nil {
				return nil, &Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("Failed to wake device with address %s", device.MACAddress)}
			}
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if err := s.writeDevice(device, add); err != nil {
			return nil, &Error{err: err, Status: http.StatusInternalServerError, Message: "Could not unmarshal JSON"}
		}
		w.WriteHeader(http.StatusNoContent)
		return nil, nil
	}
	return nil, &Error{
		Status:  http.StatusMethodNotAllowed,
		Message: fmt.Sprintf("Invalid method %s, must be %s or %s", r.Method, http.MethodGet, http.MethodPost),
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) (interface{}, *Error) {
	return nil, &Error{
		Status:  http.StatusNotFound,
		Message: "Resource not found",
	}
}

type appHandler func(http.ResponseWriter, *http.Request) (interface{}, *Error)

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, e := fn(w, r)
	if e != nil { // e is *Error, not os.Error.
		if e.err != nil {
			log.Print(e.err)
		}
		out, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}
		w.WriteHeader(e.Status)
		w.Write(out)
	} else if data != nil {
		out, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		w.Write(out)
	}
}

func requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/wake", appHandler(s.defaultHandler))
	// Return 404 in JSON for all unknown requests under /api/
	mux.Handle("/api/", appHandler(notFoundHandler))
	if s.StaticDir != "" {
		fs := http.StripPrefix("/static/", http.FileServer(http.Dir(s.StaticDir)))
		mux.Handle("/static/", fs)
	}
	return requestFilter(mux)
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
