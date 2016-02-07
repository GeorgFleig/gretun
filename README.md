# gretun
Establish GRE tunnels between server and clients

**Important**: This is an untested prototype!

## Future Plans
* both: read configuration from file
* both: test for successful tunnel creating
* both: test for existing tunnels at startup
* client: improve determination of local ip address, make sure the right one is chosen
* server: dynamically create list of client/server ip addresses using a start address and a subnet size

## Server
Manages a pool of GRE tunnel IP's, creates tunnel to clients
```
go get github.com/GeorgFleig/gretun/server
$GOPATH/bin/server
```

## Client
Requests GRE tunnel IP from server, creates tunnel to server
```
go get github.com/GeorgFleig/gretun/client
$GOPATH/bin/client <reg/unreg/destroy>
```
