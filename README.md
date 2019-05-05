# zilliqa-go
This project aims to mimic the architectural transaction flow of Zilliqa Blockchain as per this whitepaper: https://docs.zilliqa.com/whitepaper.pdf

# Design
## Nodes: 
- Config Maintainer service (who maintains nodes and their networks and serves requests)
- Shard Miners (or Normal Miners - who mines Block of transactions)
- DS Miners (who mines Block of Blocks)

## Overall Mimic Setup:
- All the nodes are started as Go Routines.
- Nodes communicate using Channels. 
- Each node have their own Listening channel - if node1 wants to communicate with node2, it uses node2's Listening channel to send messages.
- There is a common transaction channel, which is listened by all nodes. (This way, the node which is free automatically receives the transaction request first)
- When a node first receives transaction from Transaction Channel, it broadcasts to all other nodes in the network by 1. Contacting ConfigMaintainer to recieve list of other nodes in the network, 2. Sending to all nodes in list by starting separate go routines for each - to avoid blocking.


