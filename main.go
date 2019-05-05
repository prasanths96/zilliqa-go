/*
// ============================
	Project: Zilliqa-GO
	Author: Prasanth Sundaravelu
// ============================
*/

package main

import (
	"fmt"
	// "math/rand"
	// "runtime"
	"sync"
	"time"
	"runtime/trace"
	"log"
	"os"
)


// ===================================================================================
// GLOBAL CONSTS
// ===================================================================================

const NETWORKSIZE = 1
const NODESIZEPERNETWORK = 3

// ===================================================================================
// GLOBAL VARS
// ===================================================================================

var wg sync.WaitGroup
var logPrefix = "Main" + ": " + time.Now().Format("Mon Jan _2 15:04:05 2006") + ": "
var _logger = log.New(os.Stderr, logPrefix, 0)


func main() {
	// Tracer 
	_logger.Print("Starting Tracer")
	f, err := os.Create("trace.out")
    if err != nil {
		_logger.Fatalf("failed to create trace output file: %v", err)
    }
    defer func() {
        if err := f.Close(); err != nil {
            _logger.Fatalf("failed to close trace file: %v", err)
        }
	}()
	if err := trace.Start(f); err != nil {
        _logger.Fatalf("failed to start trace: %v", err)
    }
	defer trace.Stop()
	// 


	// ACTUAL PROGRAM STARTS HERE
	_logger.Print("Starting Program")
	start()
	
}

// ===================================================================================
// MAIN STARTING POINT
// ===================================================================================

func start() {
	// Start Config Maintainer 
	configChannel := startConfigMaintainer()
	tchannel := make(chan ComObj)

	wg.Add(1)
	go keepalive(configChannel)

	// Fake registration
	// wg.Add(1)
	// go fakereg(configChannel)

	// Start normal nodes
	startMiners(configChannel, tchannel)

	// Send transactions
	sendTransactions(tchannel)
	
	wg.Wait()
	_logger.Print(configChannel)
	return
}

func keepalive(ch chan<- ComObj) {
	defer wg.Done()
	heartBeat := ComObj {
		Code: KEEPALIVE,
		Data: nil,
	}
	for {
		time.Sleep(60 * time.Second)
		ch <- heartBeat
	}	
}

func fakereg(configCh chan<- ComObj) {
	ch := make(chan ComObj)
	reg := RegisterNode {
		Name: "fakeNode1",
		Network: "fakeNetwork1",
		Channel: ch,
	}

	obj := ComObj {
		Code: REGISTERNODE,
		Data: reg,
	}

	configCh <- obj
	_logger.Print("Returning from fakereg")
	return
}

func startMiners(ch chan<- ComObj, tchannel <-chan ComObj) {
	bh := BlockHeader {
		Blocknum: 0,
		PrevHash: "0",
		Nonce: 0,
	}
	genesis := Block {
		BlockHeader: bh,
		BlockHash: "DSHHBDS2F323HHDSB",
		MinerSignature: "HF3423BRYF3U328FN3283RN",
	}

	for i:=0; i<NETWORKSIZE; i++ {
		network := fmt.Sprintf("network%v", i)
		_logger.Print(network)
		for j:=0; j<NODESIZEPERNETWORK; j++ {
			name := fmt.Sprintf("node%v%v", i, j)	
			wg.Add(1)	
			go startNormalMiner(ch, genesis, tchannel, name, network)
		}
	}
}

func sendTransactions(tchannel chan<- ComObj) {

	dummyTransaction := Transaction {
		TransactionHash: "HJFSHS3437HBJKFDFSDS",
		TransactionData: "A sends 100 to B",
		InvokerSignature: "IDSBBDSDS662732DSDSGF",
	}

	request := ComObj {
		Code: SUBMITTRANSACTION,
		Data: dummyTransaction,
	}

	for i:=0; i<9; i++ {
		tchannel <- request
		_logger.Print("Transaction request returned - Success")
	}

	return
}
