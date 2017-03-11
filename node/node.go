package node

// Uses go-libp2p

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	swarm "gx/ipfs/QmNT1JbT5S89ew7kozkHoX5KUG1rfPZvTU3oMDRyJua7rU/go-libp2p-swarm"
	crypto "gx/ipfs/QmNiCwBNA8MWDADTFVq1BonUEJbS2SvjAoNkZZrhEwcuUi/go-libp2p-crypto"
	pstore "gx/ipfs/QmQMQ2RUjnaEEX8ybmrhuFFGhAwPjyL1Eo6ZoJGD7aAccM/go-libp2p-peerstore"
	net "gx/ipfs/QmRuZnMorqodado1yeTQiv1i9rmtKj29CjPSsBKM7DFXV4/go-libp2p-net"
	bhost "gx/ipfs/QmSNJRX4uphb3Eyp69uYbpRVvgqjPxfjnJmjcdMWkDH5Pn/go-libp2p/p2p/host/basic"
	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	peer "gx/ipfs/QmZcUPvPhD1Xvk6mwijYF8AfR3mG31S1YsEfHG4khrFPRr/go-libp2p-peer"
	host "gx/ipfs/QmbzbRyd22gcW92U1rA2yKagB3myMYhk45XBknJ49F9XWJ/go-libp2p-host"
	"io/ioutil"
	"log"
	"os"
)

type Node struct {
	Config *Config
	Host   host.Host
	Log    chan string
	DB     *bolt.DB
}

type Config struct {
	ListenAddr string  `json:"listen"`
	Bootstrap  []RHost `json:"bootstrap"`
	Secret     string  `json:"secret"`
}

type RHost struct {
	Peer   string `json:"peer"`
	Addr   string `json:"address"`
	Maddr  ma.Multiaddr
	PeerID peer.ID
}

func ReadCfg(path string) (*Config, error) {
	var cfg Config
	fi, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fi, &cfg)
	if err != nil {
		return nil, err
	}

	for _, r := range cfg.Bootstrap {
		r.Maddr, err = ma.NewMultiaddr(r.Addr)
		if err != nil {
			return err, nil
		}
		r.PeerID, err = peer.IDB58Decode(r.Peer)
		if err != nil {
			return err, nil
		}
	}

	return &cfg, nil
}

func Init(cfg *Config) (*Node, error) {
	node := &Node{}
	node.Log = make(chan string, 5)
	node.Config = cfg

	var err error
	node.DB, err = bolt.Open("datastore", 0600, nil)
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin(true)
	if Err != nil {
		return nil, err
	}
	_, err = tx.CreateBucketIfNotExists()
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return node, nil
}

func (n *Node) Logger() {
	logFile, err := os.Create("node.log")
	if err != nil {
		fmt.Println("Failed to create log file!")
		return
	}

	logger := log.New(logFile, "", 0)
	for msg := range n.Log {
		logger.Println(msg)
	}
}

func (n *Node) OpenHandlers() {
	n.Host.SetStreamHandler("/identify", func(s net.Stream) {
		Handler(n, s)
	})
}

func (n *Node) Start() error {
	addr, err := ma.NewMultiaddr(n.Config.ListenAddr)
	if err != nil {
		return err
	}

	ps := pstore.NewPeerstore()
	privkey, pubkey, err := getkeys()
	if err != nil {
		return err
	}

	peerID, err := peer.IDFromPublicKey(pubkey)
	if err != nil {
		return err
	}

	ps.AddPrivKey(peerID, privkey)
	ps.AddPubKey(peerID, pubkey)
	ctx := context.Background()

	netw, err := swarm.NewNetwork(ctx, []ma.Multiaddr{addr}, peerID, ps, nil)
	if err != nil {
		return err
	}

	host := bhost.New(netw)

	n.Host = host
	n.OpenHandlers()
	return nil
}

func (n *Node) Bootstrap() {
	i := 0
	for _, r := range n.Config.Bootstrap {
		err := n.Identify(r)
		if err != nil {
			i++
		}
	}
	n.Log <- fmt.Sprintf("Successfully identified to %d/%d peers.", i, len(n.Config.Bootstrap))
}

func (n *Node) Terminate() error {
	err := n.Host.Close()
	if err != nil {
		return err
	}
	return nil
}

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
