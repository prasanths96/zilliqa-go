/*
// ============================
	Project: Zilliqa-GO
	Author: Prasanth Sundaravelu
// ============================
*/
package main

// ===============================================================================
// CODES AND CONSTS
// ===============================================================================

// Channel ComObj codes
const ERROR = -2
const KEEPALIVE = -1
const REGISTERNODE = 0
const SUBMITTRANSACTION = 1
const QUERYNETWORK = 2
const QUERYNETWORKRESULT = 3

// MAX_TRANSACTION_IN_BLOCK
const MINTRANSACTION = 3


// ===============================================================================
//  Node Structs
// ===============================================================================

type Node struct {
	Name					string
	Network					string
	Ledger 					[]Block
	Blockheight 			int32
	PendingTransactions 	[]Transaction	
	ListenChannel			chan ComObj
	ConfigChannel			chan<- ComObj
	
	//DS Nodes:
	// Ledger					[]BlockOfBlocks
	// PendingBlocks			[]Block
}

type BlockOfBlocks struct {
	BoBHeader			BlockHeader
	BoBData				[]Block
	BoBHash				string
	DSCSignature		string	
}

type Block struct {
	BlockHeader			BlockHeader
	BlockData 			[]Transaction
	BlockHash			string	
	MinerSignature		string
}

type BlockHeader struct {
	Blocknum		int32
	PrevHash		string
	Nonce			int32
}

type Transaction struct {
	TransactionHash			string
	TransactionData			string
	InvokerSignature		string
}

type ComObj struct {
	Code			int
	Data 			interface{}
}

type RegisterNode struct {
	Name 			string
	Network			string
	Channel			chan<- ComObj
}

type NodeChannelMap map[string]chan<- ComObj

type ConfigMainainer struct {
	Networks		map[string]NodeChannelMap 	// Map of networks which contains Map of nodes which contains listen channel of node	
	ConfigChannel	chan ComObj
}

type QueryNetworkNodesRequest struct {
	NodeName		string
	Network 		string
	ReturnChan		chan<- ComObj
}
