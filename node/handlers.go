package node

import (
	"encoding/json"
	"fmt"
	crypto "gx/ipfs/QmNiCwBNA8MWDADTFVq1BonUEJbS2SvjAoNkZZrhEwcuUi/go-libp2p-crypto"
	pstore "gx/ipfs/QmQMQ2RUjnaEEX8ybmrhuFFGhAwPjyL1Eo6ZoJGD7aAccM/go-libp2p-peerstore"
	net "gx/ipfs/QmRuZnMorqodado1yeTQiv1i9rmtKj29CjPSsBKM7DFXV4/go-libp2p-net"
	"strings"
)

func Handler(n *Node, s net.Stream) {
	proto := strings.Split(string(s.Protocol())[1:], "/")

	// Log stream protocol type and dialer IP address.
	switch proto[0] {
	case "identify":
		err := IdentifyRemote(n, s)
		if err != nil {
			n.Log <- fmt.Sprintf("Failed to identify remote peer: %s", s.Conn().RemotePeer().Pretty())
		} else {
			n.Log <- "Remote peer successfully authenticated."
		}
	case "request":
		err := HandleRequest(n, s)
		if err != nil {
			n.Log <- fmt.Sprintf("Failed to complete remote request from peer %s", s.Conn.RemotePeer.Pretty())
		} else {
			n.Log <- fmt.Sprintf("Successfully completed remote request from peer %s", s.Conn.RemotePeer.Pretty())
		}
	default:
		n.Log <- fmt.Sprintf("Unknown protocol %s\n", proto[0][1:])
	}
}

func IdentifyRemote(n *Node, s net.Stream) error {
	// Send local node public key to remote peer.
	pkey := s.Conn().LocalPrivateKey().GetPublic().(*crypto.RsaPublicKey)
	pkb, err := crypto.MarshalRsaPublicKey(pkey)
	if err != nil {
		return err
	}
	_, err = s.Write(pkb)
	if err != nil {
		return err
	}
	n.Log <- fmt.Sprintf("Sent public key to remote peer %s", s.Conn().RemotePeer().Pretty())
	buf := make([]byte, 1024)
	i, err := s.Read(buf)
	if err != nil {
		return err
	}

	secret, err := s.Conn().LocalPrivateKey().(*crypto.RsaPrivateKey).Decrypt(buf[:i])
	if err != nil {
		return err
	}
	n.Log <- fmt.Sprintf("Received %s as secret from peer %s", string(secret), s.Conn().RemotePeer().Pretty())
	if string(secret) == n.Config.Secret {
		_, err = s.Write([]byte("200"))
		if err != nil {
			return err
		}
		n.Host.Peerstore().AddPubKey(s.Conn().RemotePeer(), s.Conn().RemotePublicKey())
		addmgr := &pstore.AddrManager{}
		addmgr.AddAddr(s.Conn().RemotePeer(), s.Conn().RemoteMultiaddr(), pstore.PermanentAddrTTL)
	} else {
		_, err = s.Write([]byte("400"))
		if err != nil {
			return err
		}
	}
	return nil
}

type Entry struct {
	Title       string
	Description string
	Path        string
	Age         time.Time
}

func HandleRequest(n *Node, s net.Stream) error {
	buf := make([]bytes, 1024)
	_, err = s.Read(buf)
	if err != nil {
		return err
	}

	var q []string
	err = json.Unmarshal(buf[:i], &q)
	if err != nil {
		return err
	}

	entries := make([]Entry, len(q))
	var m map[string]interface{}
	err = n.DB.Update(func(tx *bolt.Tx) error {
		for _, x := range q {
			b := tx.Bucket([]byte("index"))
			p = b.Get([]byte(x))
			var e Entry
			json.Unmarshal(p, &e)
			m[x] = e
			return nil
		}
	})
	if err != nil {
		return err
	}
	results, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = s.Write(results)
	if err != nil {
		return err
	}
	return nil
}
