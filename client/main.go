package main

/*
GRE Tunnel Client

Requests GRE tunnel IP from server and establishes a tunnel
*/

import (
	"errors"
	"github.com/GeorgFleig/gretun/gretun"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

const serverIp = "10.5.0.18"

func main() {
	log.Println("GRE Tunnel Client")
	if len(os.Args) == 2 {
		localIp, err := getLocalIp()
		if err == nil {
			log.Printf("Using %s as local IP address.", localIp)
			tun := &gretun.Tunnel{GreNum: 1, LocalIp: localIp, RemoteIp: net.ParseIP(serverIp)}
			switch os.Args[1] {
			case "reg":
				register(tun)
			case "unreg":
				unregister(tun)
			case "destroy":
				destroy(tun)
			default:
				log.Fatalln("Method unknown. Usage: ./client <reg/unreg/destroy>")
			}
		} else {
			log.Fatalln("Cannot get local IP address!", err)
		}
	} else {
		log.Fatalln("Argument missing. Usage: ./client <reg/unreg/destroy>")
	}
}

// determines local IP address, stops at first result ..
func getLocalIp() (net.IP, error) {
	netifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, netif := range netifs {
		addresses, err := netif.Addrs()
		if err != nil {
			return nil, err
		}
		for _, address := range addresses {
			ipnet, ok := address.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("cannot find local IP address")
}

// register at server and create GRE tunnel
func register(tun *gretun.Tunnel) {
	log.Println("Requesting GRE IP from server ..")
	var greIp net.IP = nil
	statusCode, body := httpReq("http://" + serverIp + ":8080/reg")
	if statusCode != 0 {
		log.Printf("Response: %d (%s): %s\n", statusCode, http.StatusText(statusCode), body)
		greIp = net.ParseIP(string(body))
	}
	if greIp != nil {
		tun.CGreIp = greIp
		if tun.Create() {
			log.Println("GRE Tunnel created.")
		} else {
			log.Println("Failed creating GRE Tunnel.")
		}
	} else {
		log.Fatalln("Did not get a GRE IP address from server.")
	}
}

// unregister at server and destroy GRE tunnel
func unregister(tun *gretun.Tunnel) {
	log.Println("Unregister: request termination of GRE tunnel at server ..")
	statusCode, body := httpReq("http://" + serverIp + ":8080/unreg")
	if statusCode == http.StatusOK {
		destroy(tun)
	} else {
		log.Printf("Response: %d (%s): %s\n", statusCode, http.StatusText(statusCode), body)
	}
}

// destroy GRE tunnel
func destroy(tun *gretun.Tunnel) {
	if tun.Destroy() {
		log.Println("GRE Tunnel destroyed.")
	} else {
		log.Println("Failed destroying GRE Tunnel.")
	}
}

// handle http GET requests
func httpReq(url string) (int, []byte) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	statusCode := res.StatusCode
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	} else {
		return statusCode, body
	}
	return 0, nil
}
