package node

// Uses go-libp2p

import (
	"context"
	"fmt"
	"gitlab.com/dubbled/dstore/config"
	crypto "gx/ipfs/QmNiCwBNA8MWDADTFVq1BonUEJbS2SvjAoNkZZrhEwcuUi/go-libp2p-crypto"
	pstore "gx/ipfs/QmQMQ2RUjnaEEX8ybmrhuFFGhAwPjyL1Eo6ZoJGD7aAccM/go-libp2p-peerstore"
	net "gx/ipfs/QmRuZnMorqodado1yeTQiv1i9rmtKj29CjPSsBKM7DFXV4/go-libp2p-net"
	bhost "gx/ipfs/QmSNJRX4uphb3Eyp69uYbpRVvgqjPxfjnJmjcdMWkDH5Pn/go-libp2p/p2p/host/basic"
	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	swarm "gx/ipfs/QmY8hduizbuACvYmL4aZQbpFeKhEQJ1Nom2jY6kv6rL8Gf/go-libp2p-swarm"
	peer "gx/ipfs/QmZcUPvPhD1Xvk6mwijYF8AfR3mG31S1YsEfHG4khrFPRr/go-libp2p-peer"
	host "gx/ipfs/QmbzbRyd22gcW92U1rA2yKagB3myMYhk45XBknJ49F9XWJ/go-libp2p-host"
	"io/ioutil"
	"log"
	"os"
)

type Node struct {
	Config *config.Config
	Host   host.Host
	Log    *log.Logger
}

var nodes []*Node

func Init(cfg *config.Config) (*Node, error) {
	logFile, err := os.Create("node.log")
	if err != nil {
		return nil, err
	}

	logger := log.New(logFile, "", 0)
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip%s/%s/tcp/%s", cfg.ListenAddr[0], cfg.ListenAddr[1], cfg.ListenAddr[2]))
	if err != nil {
		return nil, err
	}
	ps := pstore.NewPeerstore()

	// Read private key from key directory.
	privkey, pubkey, err := getkeys()
	if err != nil {
		return nil, err
	}

	peerID, err := peer.IDFromPublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	ps.AddPrivKey(peerID, privkey)
	ps.AddPubKey(peerID, pubkey)

	ctx := context.Background()

	netw, err := swarm.NewNetwork(ctx, []ma.Multiaddr{addr}, peerID, ps, nil)
	if err != nil {
		return nil, err
	}

	host := bhost.New(netw)
	node := &Node{cfg, host, logger}
	node.Log.Printf("Initialized node %s", node.Host.ID().Pretty())

	host.SetStreamHandler("/identify", func(s net.Stream) {
		Handler(node, s)

	})

	for i, remote := range cfg.Bootstrap {
		err := node.Identify(remote)
		if err != nil {
			fmt.Printf("Failed to identify to bootstrap node %d.\n", i)
			node.Log.Printf("Failed to identify to bootstrap node %d.\n", i)
		}
	}

	nodes = append(nodes, node)
	return node, nil
}

func (n *Node) bootstrap() int {
	i := 0
	for _, rhost := range n.Config.Bootstrap {
		err := n.Identify(rhost)
		if err != nil {
			continue
		} else {
			i++
		}
	}
	return i
}

func (n *Node) Terminate() error {
	err := n.Host.Close()
	if err != nil {
		return err
	}
	return nil
}

// Load or generate RSA keys for node.

func getkeys() (crypto.PrivKey, crypto.PubKey, error) {
	var privkey crypto.PrivKey
	var pubkey crypto.PubKey
	pdat, err := ioutil.ReadFile("keys/priv.key")
	if err != nil {
		privkey, pubkey, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
		if err != nil {
			return nil, nil, err
		}
		privkdata, err := crypto.MarshalPrivateKey(privkey)
		if err != nil {
			return nil, nil, err
		}
		err = ioutil.WriteFile("keys/priv.key", privkdata, os.ModePerm)
		if err != nil {
			return nil, nil, err
		}
		pubkdata, err := crypto.MarshalPublicKey(pubkey)
		if err != nil {
			return nil, nil, err
		}
		err = ioutil.WriteFile("keys/pub.key", pubkdata, os.ModePerm)
		if err != nil {
			return nil, nil, err
		}
	} else {
		privkey, _ = crypto.UnmarshalPrivateKey(pdat)
		pubkey = privkey.GetPublic()
	}
	return privkey, pubkey, nil
}
