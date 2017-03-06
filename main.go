package main

import (
	"fmt"
	n "gitlab.com/dubbled/dstore/node"
)

func main() {
	cfg, err := n.ReadCfg("node.cfg")
	if err != nil {
		panic(err)
	}
	node, err := n.Init(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Started node %s\n", node.Host.ID().Pretty())
	select {}
}
