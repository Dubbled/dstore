package node

import (
	"context"
	"fmt"
	"github.com/dubbled/dstore/node"
	swarm "gx/ipfs/QmNT1JbT5S89ew7kozkHoX5KUG1rfPZvTU3oMDRyJua7rU/go-libp2p-swarm"
	pstore "gx/ipfs/QmQMQ2RUjnaEEX8ybmrhuFFGhAwPjyL1Eo6ZoJGD7aAccM/go-libp2p-peerstore"
	bhost "gx/ipfs/QmSNJRX4uphb3Eyp69uYbpRVvgqjPxfjnJmjcdMWkDH5Pn/go-libp2p/p2p/host/basic"
	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	ltest "gx/ipfs/QmYTzt6uVtDmB5U3iYiA165DQ39xaNLjr8uuDhDtDByXYp/go-testutil"
	"strconv"
	"testing"
)

func TestNodes(t *testing.T) {
	numNodes := 10
	port := 50010
	mockRemotes := make([]node.RHost, numNodes)
	for i := 0; i < numNodes; i++ {
		cfg := &node.Config{
			fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", strconv.Itoa(port)),
			mockRemotes,
			"testing",
		}

		ident, err := ltest.RandIdentity()
		if err != nil {
			t.Error(err)
		}

		n, err := node.Init(cfg)
		if err != nil {
			t.Error(err)
		}

		addr, err := ma.NewMultiaddr(n.Config.ListenAddr)
		if err != nil {
			t.Error(err)
		}

		ps := pstore.NewPeerstore()
		ps.AddPrivKey(ident.ID(), ident.PrivateKey())
		ps.AddPubKey(ident.ID(), ident.PublicKey())

		ctx := context.Background()

		netw, err := swarm.NewNetwork(ctx, []ma.Multiaddr{addr}, ident.ID(), ps, nil)
		if err != nil {
			t.Error(err)
		}

		n.Host = bhost.New(netw)
		n.OpenHandlers()
		port++
		mockRemotes[i] = node.RHost{ident.ID().Pretty(), n.Config.ListenAddr}
	}
	for _, m := range mockRemotes {
		// if m.Addr == "" {
		//	continue
		// }
		fmt.Printf("%s: %s\n", m.Addr, m.PeerID)
	}
}
