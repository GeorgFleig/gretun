package gretun

/*
gretun lib

Provides tunnel information store and methods to create and destroy GRE tunnels
*/

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

// tunnel information store
type Tunnel struct {
	GreNum   int
	LocalIp  net.IP
	RemoteIp net.IP
	SGreIp   net.IP // server GRE IP
	CGreIp   net.IP // client GRE IP
}

// create GRE tunnel using "ip tunnel"
func (tun *Tunnel) Create() bool {
	greIp := tun.CGreIp
	if tun.SGreIp != nil { // detect execution on server
		greIp = tun.SGreIp
	}
	log.Printf("Tunnel: GreNum: %d, LocalIp: %s, RemoteIp: %s SGreIp: %s, CGreIp: %s\n", tun.GreNum, tun.LocalIp, tun.RemoteIp, tun.SGreIp, tun.CGreIp)
	command := fmt.Sprintf("/usr/bin/sudo /bin/sh -c 'ip tunnel add gre%d mode gre remote %s local %s ttl 255 && ip link set gre%d up && ip addr add %s/30 dev gre%d'", tun.GreNum, tun.RemoteIp, tun.LocalIp, tun.GreNum, greIp, tun.GreNum)
	return execCmd(command)
}

// destroy GRE tunnel using "ip tunnel"
func (tun *Tunnel) Destroy() bool {
	log.Printf("Destroying tunnel gre%d\n", tun.GreNum)
	command := fmt.Sprintf("/usr/bin/sudo /bin/sh -c 'ip link set gre%d down && ip tunnel del gre%d'", tun.GreNum, tun.GreNum)
	return execCmd(command)
}

// execution helper
func execCmd(cmd string) bool {
	log.Println("Exec:", cmd)
	exec := exec.Command(strings.Fields(cmd)[0], strings.Fields(cmd)[1:]...)
	err := exec.Run()
	if err != nil {
		return false
	}
	return true
}
