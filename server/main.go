package main

import (
	"fmt"
	"github.com/GeorgFleig/gretun/gretun"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const serverIp = "10.5.0.18"

func main() {
	// list of gre ip addresses, in this case /30 subnets are used for each tunnel
	sGreIpList := []string{"192.168.0.1", "192.168.0.5", "192.168.0.9"}  // input: list of server GRE IPs
	cGreIpList := []string{"192.168.0.2", "192.168.0.6", "192.168.0.10"} // input: list of client GRE IPs
	tunList := make(TunList, 0, len(sGreIpList))                         // prepare tunnel collection
	for i, ip := range sGreIpList {
		tunList = append(tunList, gretun.Tunnel{GreNum: i + 1, LocalIp: net.ParseIP(serverIp), SGreIp: net.ParseIP(ip), CGreIp: net.ParseIP(cGreIpList[i])})
	}

	// handle interrupt to destroy all tunnels before exiting
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println("Caught signal:", sig)
		tunList.destroyAll()
		os.Exit(0)
	}()

	// start webserver
	log.Println("GRE Tunnel Server running, waiting for HTTP requests ..")
	log.Printf("IP pool size: %d\n", len(sGreIpList))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { reqHandler(w, r, &tunList) }) // handle requests on /
	http.ListenAndServe(":8080", nil)
}

type TunList []gretun.Tunnel

// handle incoming HTTP requests
func reqHandler(w http.ResponseWriter, r *http.Request, tunList *TunList) {
	log.Printf("Incoming HTTP request: %s\n", r.URL.Path)
	remoteIpTmp, _, _ := net.SplitHostPort(r.RemoteAddr)
	remoteIp := net.ParseIP(remoteIpTmp)
	action := ""
	if len(r.URL.Path) > 1 {
		action = strings.Split(r.URL.Path, "/")[1]
	}
	switch action {
	case "reg":
		log.Printf("Client %s wants to register.\n", remoteIp)
		registerClient(w, r, remoteIp, tunList)
	case "unreg":
		log.Printf("Client %s wants to unregister.\n", remoteIp)
		unregisterClient(w, r, remoteIp, tunList)
	default:
		log.Printf("Cannot handle HTTP request from %s: %s\n", remoteIp, r.URL.Path)
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// register client at server and create GRE tunnel
func registerClient(w http.ResponseWriter, r *http.Request, remoteIp net.IP, tunList *TunList) {
	isReg, tun := tunList.get(&remoteIp, true)
	if !isReg {
		if tun != nil {
			if tun.Create() {
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, "%s", (*tun).CGreIp) // send GRE IP address
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Cannot create tunnel!")
				log.Printf("ERROR: Cannot create tunnel.\n")
			}
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "No IP addresses left, sorry!")
			log.Printf("ERROR: No IP addresses left.\n")
		}
	} else {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "There is already a GRE tunnel allocated for %s. Unregister first.", remoteIp)
		log.Printf("ERROR: %s tried to register but tunnel already exists.\n", remoteIp)
	}
}

// unregister client at server and destroy GRE tunnel
func unregisterClient(w http.ResponseWriter, r *http.Request, remoteIp net.IP, tunList *TunList) {
	isReg, tun := tunList.get(&remoteIp, false)
	if isReg { // is there an existing GRE tunnel?
		if tun.Destroy() {
			tunList.free(tun) // free allocated GRE ip, put this in destroy? only needed on server..
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Unregistered.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Cannot destroy tunnel!")
			log.Printf("ERROR: Cannot destroy tunnel.\n")
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No GRE Tunnel allocated for %s.", remoteIp)
		log.Printf("ERROR: %s tried to unregister but no tunnel was found.\n", remoteIp)
	}
}

// handles creation of new and lookup of existing tunnels
// retuns: isRegistered (bool), tunnel *Tunnel
func (tunList *TunList) get(remoteIp *net.IP, createTun bool) (bool, *gretun.Tunnel) {
	for i := range *tunList {
		tun := &(*tunList)[i] // a pointer is used to refer to the existing tunnel
		if createTun {
			isReg, tmpTun := tunList.get(remoteIp, false)
			if isReg {
				return true, tmpTun // client is already registered, cannot create new tunnel
			} else if tun.RemoteIp == nil {
				tun.RemoteIp = *remoteIp
				return false, tun // client not registered, new tunnel created
			} else {
				return false, nil // client not registered, but ip pool is empty
			}
		} else if tun.RemoteIp.Equal(*remoteIp) {
			return true, tun // client is registered, no new tunnel requested
		}
	}
	return false, nil // client is not registered, no new tunnel requested
}

// unsets remote IP to free GRE IP
func (tunList *TunList) free(tun *gretun.Tunnel) {
	tun.RemoteIp = nil
}

// destroys all tunnels
func (tunList *TunList) destroyAll() {
	log.Println("Destroying all remaining tunnels ..")
	for i := range *tunList {
		tun := &(*tunList)[i]
		if tun.RemoteIp != nil {
			tun.Destroy()
		}
	}
}
