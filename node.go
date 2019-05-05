/*
// ===================================================
	Project: Zilliqa-GO
	Author: Prasanth Sundaravelu
// ====================================================
*/

package main

import (
	// "fmt"
	// "math/rand"
	// "runtime"
	// "sync"
	"time"
	"log"
	"errors"
	"os"
	"math/rand"
)

// ===============================================================================
// Config Maintainer
// ===============================================================================


func startConfigMaintainer() chan<- ComObj{
	var logPrefix = "ConfigMaintainer" + ": " + time.Now().Format("Mon Jan _2 15:04:05 2006") + ": "
	var _logger = log.New(os.Stderr, logPrefix, 0)

	_logger.Print("Initializing Node")

	channel := make(chan ComObj)

	config := ConfigMainainer {
		Networks: make(map[string]NodeChannelMap),
		ConfigChannel: channel,
	}

	_logger.Print("Starting Node")
	wg.Add(1)
	go config.configMaintainerRoutine()

	return channel
}

func (c *ConfigMainainer) configMaintainerRoutine() {
	var logPrefix = "ConfigMaintainerRoutine" + ": " + time.Now().Format("Mon Jan _2 15:04:05 2006") + ": "
	var _logger = log.New(os.Stderr, logPrefix, 0)

	// defer close(c.ConfigChannel)
	defer wg.Done()
	// Listen to channel
	_logger.Print("Started Listening")
	defer _logger.Print("Closing...")
	for {
		select {
		case d, ok := <-c.ConfigChannel:
			if !ok {
				_logger.Print("Problem with ConfigChannel.")
				return 
			}
			switch d.Code {
			case REGISTERNODE:
				_logger.Print("Registering Node")
				nodeinfo := d.Data.(RegisterNode)
				err := c.registerNode(nodeinfo)  
				if err != nil {
					_logger.Print("Error: " + err.Error())
				} else {
					_logger.Print("Node registered successfully")
					_logger.Print(c)
				}
			case QUERYNETWORK:
				request := d.Data.(QueryNetworkNodesRequest)
				_logger.Print("Recieved QueryNetworkNodes Request from: " + request.NodeName)
				nodeChanMap, ok := c.Networks[request.Network]
				if !ok {
					_logger.Print("Network: " + request.Network + " is not registered.")
					failed := ComObj {
						Code: ERROR,
						Data: "Network: " + request.Network + " is not registered.",
					}
					request.ReturnChan <- failed
				}

				success := ComObj {
					Code: QUERYNETWORKRESULT,
					Data: nodeChanMap,
				}

				_logger.Print("Query Success - Returning Result to node: " + request.NodeName)
				request.ReturnChan <- success

			}
		}
	}

	// Take action
		// 1. Register
		// 2. Return network
}

func (c *ConfigMainainer) registerNode(info RegisterNode) error {
	// Check If node already exists in network 
	if _, ok := c.Networks[info.Network][info.Name]; ok {
		return errors.New("Node: " + info.Name + " already exists in network: " + info.Network)
	}
	// Check if network is not present yet
	if _, ok := c.Networks[info.Network]; !ok {
		c.Networks[info.Network] = make(NodeChannelMap)
	}

	c.Networks[info.Network][info.Name] = info.Channel

	return nil
}


// ===============================================================================
// Normal Miner
// ===============================================================================


func startNormalMiner(configChannel chan<- ComObj, genesis Block, transactionChannel <-chan ComObj, name string, network string){
	var logPrefix = name + ": " + time.Now().Format("Mon Jan _2 15:04:05 2006") + ": "
	var _logger = log.New(os.Stderr, logPrefix, 0)

	_logger.Print("Initializing Node")

	nodeInfo := Node {
		Name: name,
		Network: network,		
		ListenChannel: make(chan ComObj),
		Ledger: []Block{genesis},
		Blockheight: 0,
		ConfigChannel: configChannel,
	}

	_logger.Print("Registering with ConfigChannel")
	initConfig := RegisterNode {
		Name: name,
		Network: network,
		Channel: nodeInfo.ListenChannel,
	}
	comObj := ComObj {
		Code: REGISTERNODE,
		Data: initConfig,
	}

	configChannel <- comObj
	go nodeInfo.minerRoutine(transactionChannel)		
}

func (n *Node) minerRoutine(transactionChannel <-chan ComObj) {
	var logPrefix = "MinerRoutine-" + n.Name + ": " + time.Now().Format("Mon Jan _2 15:04:05 2006") + ": "
	var _logger = log.New(os.Stderr, logPrefix, 0)

	defer wg.Done()

	_logger.Print("Listening to Transaction channel and ListenChannel")
	
	for {
		select {
		case d, ok := <- transactionChannel:
			if !ok {
				_logger.Print("Problem with TransactionChannel.")
				return 
			}
			switch d.Code {
			case SUBMITTRANSACTION: 
				// Broadcast must happend only when received in transactionChannel
				_logger.Print("Broadcasting transaction to other miners.")
				n.broadcastTransaction(d, _logger)
				n.minerSubmitTransactionProcessor(d, _logger)
			}
		case d, ok := <- n.ListenChannel: 
		if !ok {
			_logger.Print("Problem with TransactionChannel.")
			return 
		}
		// _logger.Print("ListenChannel Received: ", d)
		switch d.Code {
			case SUBMITTRANSACTION: 
				n.minerSubmitTransactionProcessor(d, _logger)
			}
		}
	}

}

func (n *Node) broadcastTransaction(comObj ComObj, _logger *log.Logger) {
	_logger.Print("Requesting channels of nodes in this network to ConfigMaintainer.")
	tempChan := make(chan ComObj)
	defer close(tempChan)

	queryData := QueryNetworkNodesRequest {
		NodeName: n.Name,
		Network: n.Network,
		ReturnChan: tempChan,
	}
	queryRequest := ComObj {
		Code: QUERYNETWORK,
		Data: queryData,
	}
	n.ConfigChannel <- queryRequest

	result := <-tempChan
	switch result.Code {
	case ERROR:
		_logger.Print("Received Error: " + result.Data.(string))
		return
	case QUERYNETWORKRESULT:
		_logger.Print("Recieved Channels of Nodes in this network")
		nodeChanMap := result.Data.(NodeChannelMap)
		_logger.Print(nodeChanMap)
		// Deleting current node from list
		delete(nodeChanMap, n.Name)
		_logger.Print("Broadcasting Transaction")
		send(comObj, nodeChanMap, _logger)
	}

}

func send (comObj ComObj, nodeChanMap NodeChannelMap, _logger *log.Logger) {
	for key := range nodeChanMap {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_logger.Print("Sending Transaction to: " + name)
			nodeChanMap[name] <- comObj
		}(key)
	}
	return
}

func (n *Node) minerSubmitTransactionProcessor(d ComObj, _logger *log.Logger) {
	_logger.Print("Received a Transaction.")
	transaction := d.Data.(Transaction)
	
	_logger.Print("Added Transaction to PendingTransactions.")
	n.PendingTransactions = append(n.PendingTransactions, transaction)
	_logger.Print(n.PendingTransactions)

	if len(n.PendingTransactions) >= MINTRANSACTION {
		_logger.Print("Mining Block #", (n.Blockheight + 1))
		n.mineBlock(_logger)
		_logger.Print("Mining successful.")
		// _logger.Print("Starting consensus.")
	} 
}

func (n *Node) mineBlock(_logger *log.Logger) {
	blockHeader := BlockHeader {
		Blocknum: int32(len(n.Ledger)),
		PrevHash: n.Ledger[len(n.Ledger)-1].BlockHash,
		Nonce: int32(rand.Intn(1000)),
	}

	block := Block {
		BlockHeader: blockHeader,
		BlockData: n.PendingTransactions,
		BlockHash: "JMFDSB6537HHDSBH372",
		MinerSignature: "HFDNF46343DSDGS",
	}

	_logger.Print("Added new block: ")
	_logger.Print(block)

	n.Ledger = append(n.Ledger, block)
	n.PendingTransactions = make([]Transaction, 0)
	n.Blockheight = n.Blockheight + 1

	return 
}

// ===============================================================================
// DS Miner
// ===============================================================================


// func startDSMiner(netChannel chan []byte, shardDSChannel chan []byte) {
// }

