# dstore
Distributed storage system based on go-libp2p.

### Planned Features
1. Streaming data in chunks to clients.
2. Bootstraping nodes from config.
3. Different protocols to handle updating, distribution, and syncing of peerstores.
4. Peerstore dump so that nodes do not have to re-identify to other nodes on the network.
5. Salted secret verification to confirm authentication to join network.
6. Algorithm to choose viable set of nodes for a client to utilize.
