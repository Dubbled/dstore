package node

import (
	"context"
	"errors"
	"gitlab.com/dubbled/dstore/config"
	crypto "gx/ipfs/QmNiCwBNA8MWDADTFVq1BonUEJbS2SvjAoNkZZrhEwcuUi/go-libp2p-crypto"
	pstore "gx/ipfs/QmQMQ2RUjnaEEX8ybmrhuFFGhAwPjyL1Eo6ZoJGD7aAccM/go-libp2p-peerstore"
	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	peer "gx/ipfs/QmZcUPvPhD1Xvk6mwijYF8AfR3mG31S1YsEfHG4khrFPRr/go-libp2p-peer"
)

func (n *Node) Identify(target config.RHost) error {
	ctx := context.Background()
	maddr, err := ma.NewMultiaddr(target.Addr)
	if err != nil {
		return err
	}
	peerID, err := peer.IDFromString(target.PeerID)
	if err != nil {
		return err
	}
	n.Host.Peerstore().AddAddr(peerID, maddr, pstore.PermanentAddrTTL)
	s, err := n.Host.NewStream(ctx, peerID, "/identify")
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)
	i, err := s.Read(buf)
	if err != nil {
		return err
	}

	// Get public key from buffer.
	rkey, err := crypto.UnmarshalRsaPublicKey(buf[:i])
	if err != nil {
		return err
	}

	encSec, err := rkey.Encrypt([]byte(n.Config.Secret + peerID.Pretty())) // Encrypt secret with salt as Peer ID
	if err != nil {
		return err
	}
	_, err = s.Write(encSec)
	if err != nil {
		return err
	}
	buf = make([]byte, 1024)
	i, err = s.Read(buf)
	respCode := string(buf[:i])
	if respCode != "200" {
		return errors.New("Failed to identify to peer")
	}
	return nil
}
