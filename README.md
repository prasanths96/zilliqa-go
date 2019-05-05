# zilliqa-go
This project aims to mimic the architectural transaction flow of Zilliqa Blockchain as per this whitepaper: https://docs.zilliqa.com/whitepaper.pdf

# Design
## Nodes: 
- Config Maintainer service (who maintains nodes and their networks and serves requests)
- Shard Miners (or Normal Miners - who mines Block of transactions)
- DS Miners (who mines Block of Blocks)

