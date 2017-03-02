package main

import (
	"fmt"
	"gitlab.com/dubbled/dstore/config"
	n "gitlab.com/dubbled/dstore/node"
)

func main() {
	cfg, err := config.ReadCfg("node.cfg")
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
