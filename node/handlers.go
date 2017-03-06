package node

import (
	net "gx/ipfs/QmRuZnMorqodado1yeTQiv1i9rmtKj29CjPSsBKM7DFXV4/go-libp2p-net"
	// ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	crypto "gx/ipfs/QmNiCwBNA8MWDADTFVq1BonUEJbS2SvjAoNkZZrhEwcuUi/go-libp2p-crypto"
	"log"
	"strings"
)

func Handler(n *Node, s net.Stream) {
	proto := strings.Split(string(s.Protocol())[1:], "/")

	// Log stream protocol type and dialer IP address.
	switch proto[0] {
	case "identify":
		err := IdentifyRemote(n, s)
		if err != nil {
			log.Printf("Failed to identify remote peer: %s", s.Conn().RemotePeer().Pretty())
		}
	default:
		n.Log.Printf("Unknown protocol %s\n", proto[0][1:])
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
	n.Log.Printf("Sent public key to remote peer.\n")
	buf := make([]byte, 1024)
	i, err := s.Read(buf)
	if err != nil {
		return err
	}

	secret, err := s.Conn().LocalPrivateKey().(*crypto.RsaPrivateKey).Decrypt(buf[:i])
	if err != nil {
		return err
	}
	n.Log.Printf("Received %s as secret from peer.", string(secret))
	if string(secret) == n.Config.Secret {
		_, err = s.Write([]byte("200"))
		if err != nil {
			return err
		}
		n.Log.Printf("Remote peer successfully authenticated.\n")
	} else {
		_, err = s.Write([]byte("400"))
		if err != nil {
			return err
		}
		n.Log.Printf("Failed to authenticate remote peer.\n")
	}
	return nil
}
