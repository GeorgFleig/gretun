# gretun
Establish GRE tunnels between server and clients

**Important**: This is an untested prototype!

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
