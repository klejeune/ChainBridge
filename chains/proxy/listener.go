// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package proxy

import (
//	"context"
	"errors"
	"fmt"
	"math/big"
	"time"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/ChainSafe/ChainBridge/bindings/Bridge"
	"github.com/ChainSafe/ChainBridge/bindings/ERC20Handler"
	"github.com/ChainSafe/ChainBridge/bindings/ERC721Handler"
	"github.com/ChainSafe/ChainBridge/bindings/GenericHandler"
	"github.com/ChainSafe/ChainBridge/chains"
	utils "github.com/ChainSafe/ChainBridge/shared/ethereum"
	"github.com/ChainSafe/chainbridge-utils/blockstore"
	metrics "github.com/ChainSafe/chainbridge-utils/metrics/types"
	"github.com/ChainSafe/chainbridge-utils/msg"
	"github.com/ChainSafe/log15"
	eth "github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

var BlockRetryInterval = time.Second * 5
var BlockRetryLimit = 5
var ErrFatalPolling = errors.New("listener block polling failed")

type listener struct {
	cfg                    Config
	conn                   Connection
	router                 chains.Router
	bridgeContract         *Bridge.Bridge // instance of bound bridge contract
	erc20HandlerContract   *ERC20Handler.ERC20Handler
	erc721HandlerContract  *ERC721Handler.ERC721Handler
	genericHandlerContract *GenericHandler.GenericHandler
	log                    log15.Logger
	blockstore             blockstore.Blockstorer
	stop                   <-chan int
	sysErr                 chan<- error // Reports fatal error to core
	latestBlock            metrics.LatestBlock
	metrics                *metrics.ChainMetrics
	blockConfirmations     *big.Int
}

// NewListener creates and returns a listener
func NewListener(cfg *Config, log log15.Logger, stop <-chan int, sysErr chan<- error, m *metrics.ChainMetrics) *listener {
	return &listener{
		cfg:                *cfg,
		//conn:               conn,
		log:                log,
		//blockstore:         bs,
		stop:               stop,
		sysErr:             sysErr,
		latestBlock:        metrics.LatestBlock{LastUpdated: time.Now()},
		metrics:            m,
		//blockConfirmations: cfg.blockConfirmations,
	}
}

// // setContracts sets the listener with the appropriate contracts
// func (l *listener) setContracts(bridge *Bridge.Bridge, erc20Handler *ERC20Handler.ERC20Handler, erc721Handler *ERC721Handler.ERC721Handler, genericHandler *GenericHandler.GenericHandler) {
// 	l.bridgeContract = bridge
// 	l.erc20HandlerContract = erc20Handler
// 	l.erc721HandlerContract = erc721Handler
// 	l.genericHandlerContract = genericHandler
// }

// sets the router
func (l *listener) setRouter(r chains.Router) {
	l.router = r
}

// start registers all subscriptions provided by the config
func (l *listener) start() error {
	l.log.Debug("Starting listener...")

	go func() {
		err := l.pollBlocks()
		if err != nil {
			l.log.Error("Polling blocks failed", "err", err)
		}
	}()

	return nil
}

// pollBlocks will poll for the latest block and proceed to parse the associated events as it sees new blocks.
// Polling begins at the block defined in `l.cfg.startBlock`. Failed attempts to fetch the latest block or parse
// a block will be retried up to BlockRetryLimit times before continuing to the next block.
func (l *listener) pollBlocks() error {
	l.log.Info("Polling Blocks...")
	//var currentBlock = l.cfg.startBlock
	var retry = BlockRetryLimit

	var firstTime bool = true

	for {
		select {
		case <-l.stop:
			return errors.New("polling terminated")
		default:
			// No more retries, goto next block
			if retry == 0 {
				l.log.Error("Polling failed, retries exceeded")
				l.sysErr <- ErrFatalPolling
				return nil
			}

			if (firstTime) {
				response, err := http.Get("http://localhost:3000/deposits")
				if err != nil {
					l.log.Error("Unable to get deposits", "err", err)
					retry--
					time.Sleep(BlockRetryInterval)
					continue
				}

				responseData, err := ioutil.ReadAll(response.Body)
				if err != nil {
					l.log.Error("Unable to read deposit response", "err", err)
					retry--
					time.Sleep(BlockRetryInterval)
					continue
				}

				fmt.Println(string(responseData))

				var depositsResponse DepositResponse
				json.Unmarshal(responseData, &depositsResponse)

				fmt.Println("Found", len(depositsResponse.Deposits), " deposits")
				// fmt.Println(depositsResponse.Date)
				// fmt.Println(depositsResponse.Deposits[0].From)
				// fmt.Println(depositsResponse.Deposits[0].To)
				// fmt.Println(depositsResponse.Deposits[0].Amount)
				// fmt.Println(depositsResponse.Deposits[0].Currency)
				// fmt.Println(depositsResponse.Deposits[0].Date)
				// fmt.Println(depositsResponse.Deposits[0].Status)

				time.Sleep(BlockRetryInterval)

				// latestBlock, err := l.conn.LatestBlock()
				// if err != nil {
				// 	l.log.Error("Unable to get latest block", "block", currentBlock, "err", err)
				// 	retry--
				// 	time.Sleep(BlockRetryInterval)
				// 	continue
				// }

				// if l.metrics != nil {
				// 	l.metrics.LatestKnownBlock.Set(float64(latestBlock.Int64()))
				// }

				// // Sleep if the difference is less than BlockDelay; (latest - current) < BlockDelay
				// if big.NewInt(0).Sub(latestBlock, currentBlock).Cmp(l.blockConfirmations) == -1 {
				// 	l.log.Debug("Block not ready, will retry", "target", currentBlock, "latest", latestBlock)
				// 	time.Sleep(BlockRetryInterval)
				// 	continue
				// }

				// Parse out events
				err = l.getDepositEventsForBlock(depositsResponse.Deposits)
				if err != nil {
					l.log.Error("Failed to get events for block", "err", err)
					retry--
					continue
				}

				firstTime = false
			}

			// // Write to block store. Not a critical operation, no need to retry
			// err = l.blockstore.StoreBlock(currentBlock)
			// if err != nil {
			// 	l.log.Error("Failed to write latest block to blockstore", "block", currentBlock, "err", err)
			// }

			// if l.metrics != nil {
			// 	l.metrics.BlocksProcessed.Inc()
			// 	l.metrics.LatestProcessedBlock.Set(float64(latestBlock.Int64()))
			// }

			// l.latestBlock.Height = big.NewInt(0).Set(latestBlock)
			// l.latestBlock.LastUpdated = time.Now()

			// Goto next block and reset retry counter
			//currentBlock.Add(currentBlock, big.NewInt(1))
			//retry = BlockRetryLimit
		}
	}
}

// getDepositEventsForBlock looks for the deposit event in the latest block
func (l *listener) getDepositEventsForBlock(deposits []Deposit) error {
	l.log.Debug("Processing deposits")

	// read through the log events and handle their deposit event if handler is recognized
	for _, deposit := range deposits {
		var m msg.Message

		// intDestId := new(big.Int)
		// intDestId.SetString(deposit.DestinationChainId, 10)	

		// l.log.Info("####### intDestId:", intDestId)

		// intNonce := new(big.Int)
		// intNonce.SetString(deposit.Id, 10)	

		
		// destId := msg.ChainId(intDestId.Uint64())
		// nonce := msg.Nonce(intNonce.Uint64())
		intDestId, err := strconv.ParseInt(deposit.DestinationChainId, 10, 64)
		intNonce, err := strconv.ParseInt(deposit.Nonce, 10, 64)

		destId := msg.ChainId(intDestId)
		nonce := msg.Nonce(intNonce)

		//var err error
		// addr, err := l.bridgeContract.ResourceIDToHandlerAddress(&bind.CallOpts{From: l.conn.Keypair().CommonAddress()}, rId)
		// if err != nil {
		// 	return fmt.Errorf("failed to get handler from resource ID %x", rId)
		// }

		if deposit.Type == "fungible" {
			m, err = l.handleErc20DepositedEvent(destId, nonce, deposit)
		} else if deposit.Type == "non-fungible" {
			m, err = l.handleErc721DepositedEvent(destId, nonce, deposit)
		} else if deposit.Type == "generic" {
			m, err = l.handleGenericDepositedEvent(destId, nonce, deposit)
		} else {
			l.log.Error("event has unrecognized handler", "handler", deposit.Type)
			return nil
		}

		if err != nil {
			return err
		}

		err = l.router.Send(m)
		if err != nil {
			l.log.Error("subscription error: failed to route message", "err", err)
		}
	}

	return nil
}

// buildQuery constructs a query for the bridgeContract by hashing sig to get the event topic
func buildQuery(contract ethcommon.Address, sig utils.EventSig, startBlock *big.Int, endBlock *big.Int) eth.FilterQuery {
	query := eth.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []ethcommon.Address{contract},
		Topics: [][]ethcommon.Hash{
			{sig.GetTopic()},
		},
	}
	return query
}
