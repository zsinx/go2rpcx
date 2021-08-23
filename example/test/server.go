package main

import (
	"flag"

	"github.com/smallnest/rpcx/client"
	"github.com/zsinx/go2rpcx/example/rpc"
)

var (
	addr = flag.String("addr", "localhost:8972", "server address")
)

func main() {
	flag.Parse()

	d, _ := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	rpc.ServerForUser(*addr, d)
}
