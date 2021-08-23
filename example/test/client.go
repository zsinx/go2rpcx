package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/smallnest/rpcx/client"
	"github.com/zsinx/go2rpcx/example/rpc"
)

var (
	addr = flag.String("addr", "localhost:8972", "server address")
)

func main() {
	flag.Parse()

	d, _ := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	xClient, _ := rpc.NewXClientForUser(d)
	userRpc := rpc.NewUserClient(xClient)
	defer xClient.Close()

	args := &rpc.Request{
		Name: "Arthur",
	}

	for {
		reply, err := userRpc.GetUser(context.Background(), args)
		if err != nil {
			log.Fatalf("failed to call: %v", err)
		}

		log.Printf("args: %v, reply: %v", args, reply)
		time.Sleep(1e9)
	}
}
