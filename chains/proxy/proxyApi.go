package proxy

type Deposit struct {
	Id													string							`json:"id"`
	Nonce												string							`json:"nonce"`
	From												string							`json:"from"`
	DestinationChainId					string							`json:"destinationChainId"`
	DestinationRecipientAddress	string							`json:"destinationRecipientAddress"`
	Type												string							`json:"type"`
	ResourceId									string							`json:"resourceId"`
	Fungible										FungibleDeposit			`json:"fungible"`
	NonFungible									NonFungibleDeposit	`json:"nonFungible"`
	Generic											GenericDeposit			`json:"generic"`
	Date												string							`json:"date"`
	Status											string							`json:"status"`
}

type FungibleDeposit struct {
	Amount											string							`json:"amount"`
}

type NonFungibleDeposit struct {
	TokenId											string							`json:"tokenId"`
	Metadata										string							`json:"metadata"`
}

type GenericDeposit struct {
	Metadata										string							`json:"metadata"`
}

type DepositResponse struct {
	Deposits										[]Deposit						`json:"deposits"`
	Date												string							`json:"date"`
}