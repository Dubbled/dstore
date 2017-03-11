package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	crypto "gx/ipfs/QmNiCwBNA8MWDADTFVq1BonUEJbS2SvjAoNkZZrhEwcuUi/go-libp2p-crypto"
	pstore "gx/ipfs/QmQMQ2RUjnaEEX8ybmrhuFFGhAwPjyL1Eo6ZoJGD7aAccM/go-libp2p-peerstore"
	ma "gx/ipfs/QmSWLfmj5frN9xVLMMN846dMDriy5wN5jeghUm7aTW3DAG/go-multiaddr"
	peer "gx/ipfs/QmZcUPvPhD1Xvk6mwijYF8AfR3mG31S1YsEfHG4khrFPRr/go-libp2p-peer"
)

func (n *Node) Identify(target RHost) error {
	ctx := context.Background()
	maddr, err := ma.NewMultiaddr(target.Addr)
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

	n.Log <- fmt.Sprintf("Got public key from remote peer %s", peerID.Pretty())
	encSec, err := rkey.Encrypt([]byte(n.Config.Secret))
	if err != nil {
		return err
	}

	_, err = s.Write(encSec)
	if err != nil {
		return err
	}

	n.Log <- fmt.Sprintf("Sending encrypted secret to peer %s", peerID.Pretty())

	buf = make([]byte, 1024)
	i, err = s.Read(buf)
	respCode := string(buf[:i])
	if respCode == "200" {
		n.Log <- fmt.Sprintf("Successfully identified to peer %s", peerID.Pretty())
		n.Host.Peerstore().AddPubKey(peerID, rkey)
		return nil
	}

	n.Log <- fmt.Sprintf("Error: %s: Failed to identify to peer %s", respCode, peerID.Pretty())
	return errors.New(respCode)
}

func (n *Node) Request(id peer.ID, addr ma.Multiaddr, query []string) (map[string]interface{}, error) {
	ctx := context.Background()
	s, err := n.Host.NewStream(ctx, id, "/request")
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	_, err = s.Write(data)
	if err != nil {
		return nil, err
	}

	resp := make([]byte, 8192)
	i, err := s.Read(resp)
	if err != nil {
		return nil, err
	}
}
