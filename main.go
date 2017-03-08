package main

import (
	"fmt"
	n "github.com/dubbled/dstore/node"
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

	node.Start()
	fmt.Printf("Started node %s\n", node.Host.ID().Pretty())

	for i, remote := range cfg.Bootstrap {
		err := node.Identify(remote)
		if err != nil {
			node.Log.Printf("Failed to identify to bootstrap node %d.", i)
		}
	}

	select {}
}
