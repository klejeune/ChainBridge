// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package proxy

import (
//	"errors"
	"fmt"
	"math/big"

	utils "github.com/ChainSafe/ChainBridge/shared/ethereum"
	"github.com/ChainSafe/chainbridge-utils/core"
	"github.com/ChainSafe/chainbridge-utils/msg"
	"github.com/ethereum/go-ethereum/common"
)

const DefaultGasLimit = 6721975
const DefaultGasPrice = 20000000000
const DefaultBlockConfirmations = 10
const DefaultGasMultiplier = 1

// Chain specific options
var (
	Test                  = "test"
)

// Config encapsulates all necessary parameters in ethereum compatible forms
type Config struct {
	name                   string      // Human-readable chain name
	id                     msg.ChainId // ChainID
	endpoint               string      // url for rpc endpoint
	from                   string      // address of key to use
	keystorePath           string      // Location of keyfiles
	blockstorePath         string
	freshStart             bool // Disables loading from blockstore at start
	bridgeContract         common.Address
	erc20HandlerContract   common.Address
	erc721HandlerContract  common.Address
	genericHandlerContract common.Address
	gasLimit               *big.Int
	maxGasPrice            *big.Int
	gasMultiplier          *big.Float
	http                   bool // Config for type of connection
	startBlock             *big.Int
	blockConfirmations     *big.Int
}

// parseChainConfig uses a core.ChainConfig to construct a corresponding Config
func parseChainConfig(chainCfg *core.ChainConfig) (*Config, error) {

	config := &Config{
		name:                   chainCfg.Name,
		id:                     chainCfg.Id,
		endpoint:               chainCfg.Endpoint,
		from:                   chainCfg.From,
		keystorePath:           chainCfg.KeystorePath,
		blockstorePath:         chainCfg.BlockstorePath,
		freshStart:             chainCfg.FreshStart,
		bridgeContract:         utils.ZeroAddress,
		erc20HandlerContract:   utils.ZeroAddress,
		erc721HandlerContract:  utils.ZeroAddress,
		genericHandlerContract: utils.ZeroAddress,
		gasLimit:               big.NewInt(DefaultGasLimit),
		maxGasPrice:            big.NewInt(DefaultGasPrice),
		gasMultiplier:          big.NewFloat(DefaultGasMultiplier),
		http:                   false,
		startBlock:             big.NewInt(0),
		blockConfirmations:     big.NewInt(0),
	}

	if test, ok := chainCfg.Opts[Test]; ok && test != "" {		
		delete(chainCfg.Opts, Test)
	} else {
		return nil, fmt.Errorf("unable to parse %s", Test)
	}

	if len(chainCfg.Opts) != 0 {
		return nil, fmt.Errorf("unknown Opts Encountered: %#v", chainCfg.Opts)
	}

	return config, nil
}
