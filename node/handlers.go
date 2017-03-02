package node

import (
	net "gx/ipfs/QmRuZnMorqodado1yeTQiv1i9rmtKj29CjPSsBKM7DFXV4/go-libp2p-net"
	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	"log"
	"strings"
)

func Handler(n *Node, s net.Stream) {
	proto := strings.Split(string(s.Protocol())[1:], "/")

	// Get connector's multiaddress. Simplify it.
	r := s.Conn().RemoteMultiaddr()
	maddr, err := ma.NewMultiaddr("/tcp")
	if err != nil {
		log.Printf("%#v\n", err)
	}
	r.Decapsulate(maddr)

	// Log stream protocol type and dialer IP address.
	log.Printf("Received new %s stream from.\n", proto[0], r.String()[5:])

	switch proto[0] {
	case "identify":
		err = IdentifyRemote(n, s)
		if err != nil {
			log.Printf("Failed to identify remote peer: %s", s.Conn().RemotePeer().Pretty())
		}
	default:
		log.Printf("Unknown protocol %s\n", proto[0][1:])
	}
}

func IdentifyRemote(n *Node, s net.Stream) error {
	// Send local node public key to remote peer.
	pkey, err := s.Conn().LocalPrivateKey().GetPublic().Bytes()
	if err != nil {
		return err
	}
	_, err = s.Write(pkey)
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)
	i, err := s.Read(buf)
	if err != nil {
		return err
	}

	rpid := s.Conn().RemotePeer().Pretty()
	secret := n.Config.Secret
	if string(buf[:i]) == rpid+secret { // Decrypt secret and salt.
		_, err = s.Write([]byte("200"))
		if err != nil {
			return err
		}
	} else {
		_, err = s.Write([]byte("400"))
		if err != nil {
			return err
		}
	}
	return nil
}
