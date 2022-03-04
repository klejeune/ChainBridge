// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package proxy

import (
	"errors"
	"encoding/hex"
	"github.com/ChainSafe/chainbridge-utils/msg"
	"math/big"
)

func (l *listener) handleErc20DepositedEvent(destId msg.ChainId, nonce msg.Nonce, deposit Deposit) (msg.Message, error) {
	l.log.Info("Handling fungible deposit event", "dest", destId, "deposit", deposit)

	// record, err := l.erc20HandlerContract.GetDepositRecord(&bind.CallOpts{From: l.conn.Keypair().CommonAddress()}, uint64(nonce), uint8(destId))
	// if err != nil {
	// 	l.log.Error("Error Unpacking ERC20 Deposit Record", "err", err)
	// 	return msg.Message{}, err
	// }

	amount := new(big.Int)
	amount, ok := amount.SetString(deposit.Fungible.Amount, 10)
	resourceId, err := hex.DecodeString(deposit.ResourceId[2:])
	
	if err != nil {
		l.log.Error("Error Unpacking ERC721 Deposit resource ID", "err", err)
		return msg.Message{}, err
	}

	destinationRecipientAddress, err := hex.DecodeString(deposit.DestinationRecipientAddress[2:])

	if err != nil {
		l.log.Error("Error Unpacking ERC721 Deposit DestinationRecipientAddress", "err", err)
		return msg.Message{}, err
	}

	if !ok {
		var err = errors.New("deposit.Fungible.Amount is not a Big Int")
		l.log.Error("Error Unpacking ERC20 Deposit Record amount", "err", err)
		return msg.Message{}, err
	}

	l.log.Info("NewFungibleTransfer", "DestinationRecipientAddress", deposit.DestinationRecipientAddress, "Amount", deposit.Fungible.Amount)

	return msg.NewFungibleTransfer(
		l.cfg.id,
		destId,
		nonce,
		amount,
		msg.ResourceIdFromSlice(resourceId),
		destinationRecipientAddress,
	), nil
}

func (l *listener) handleErc721DepositedEvent(destId msg.ChainId, nonce msg.Nonce, deposit Deposit) (msg.Message, error) {
	l.log.Info("Handling nonfungible deposit event")

	// record, err := l.erc721HandlerContract.GetDepositRecord(&bind.CallOpts{From: l.conn.Keypair().CommonAddress()}, uint64(nonce), uint8(destId))
	// if err != nil {
	// 	l.log.Error("Error Unpacking ERC721 Deposit Record", "err", err)
	// 	return msg.Message{}, err
	// }

	tokenId := new(big.Int)
	tokenId, ok := tokenId.SetString(deposit.NonFungible.TokenId, 10)	

	if !ok {
		var err = errors.New("deposit.NonFungible.TokenId is not a Big Int")
		l.log.Error("Error Unpacking ERC20 Deposit Record tokenId", "err", err)
		return msg.Message{}, err
	}	

	return msg.NewNonFungibleTransfer(
		l.cfg.id,
		destId,
		nonce,
		msg.ResourceIdFromSlice([]byte(deposit.ResourceId)),
		tokenId,
		[]byte(deposit.DestinationRecipientAddress),
		[]byte(deposit.NonFungible.Metadata),
	), nil
}

func (l *listener) handleGenericDepositedEvent(destId msg.ChainId, nonce msg.Nonce, deposit Deposit) (msg.Message, error) {
	l.log.Info("Handling generic deposit event")

	// record, err := l.genericHandlerContract.GetDepositRecord(&bind.CallOpts{From: l.conn.Keypair().CommonAddress()}, uint64(nonce), uint8(destId))
	// if err != nil {
	// 	l.log.Error("Error Unpacking Generic Deposit Record", "err", err)
	// 	return msg.Message{}, nil
	// }

	return msg.NewGenericTransfer(
		l.cfg.id,
		destId,
		nonce,
		msg.ResourceIdFromSlice([]byte(deposit.ResourceId)),
		[]byte(deposit.Generic.Metadata[:]),
	), nil
}
