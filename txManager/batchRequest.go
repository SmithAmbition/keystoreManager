package txManager

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-web3/dto"
	"github.com/ethereum/go-web3/providers"
)

func (tm *TxManager) CreateNewBatchSend() *providers.BatchRequest {
	return &providers.BatchRequest{}
}
func (tm *TxManager) AddGetBlockNumber(batch *providers.BatchRequest) {
	batch.AddRequest("eth_blockNumber", nil)
}
func (tm *TxManager) AddGetTransactionCount(batch *providers.BatchRequest, address string, blockNumber string) {
	params := make([]string, 2)
	params[0] = address
	params[1] = blockNumber
	batch.AddRequest("eth_getTransactionCount", params)
}
func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}
func (tm *TxManager) AddContractCallBase(batch *providers.BatchRequest, msg *ethereum.CallMsg, blockNumber *big.Int) {
	params := make([]interface{}, 2)
	params[0] = toCallArg(*msg)
	params[1] = toBlockNumArg(blockNumber)
	batch.AddRequest("eth_call", params)
}
func (tm *TxManager) AddRawTransaction(batch *providers.BatchRequest, rawTx []byte) {
	params := make([]string, 1)
	params[0] = hexutil.Encode(rawTx)
	batch.AddRequest("eth_sendRawTransaction", params)
}
func (tm *TxManager) SendBatchRequest(batch *providers.BatchRequest) ([]dto.RequestResult, error) {
	return tm.Web3.Provider.BatchSendRequest(batch)
}
func (tm *TxManager) AddContractCall(batch *providers.BatchRequest, from dto.Address, to *dto.Address, blockNumber *big.Int, contract string, funcName string, args ...interface{}) error {

	if curContract, exist := tm.Contracts.Contracts[contract]; exist {
		data, err := curContract.Pack(funcName, args...)
		if err != nil {
			fmt.Println(err)
			return err
		}
		address := common.Address(curContract.Address)
		if to != nil {
			address = common.Address(*to)
		}
		msg := &ethereum.CallMsg{
			From:     common.Address(from),
			To:       &address,
			Gas:      6750000,
			GasPrice: big.NewInt(20e9),
			Data:     data,
		}
		tm.AddContractCallBase(batch, msg, blockNumber)
		return nil
	}
	fmt.Println(contract + " is not exist!")
	return errors.New(contract + " is not exist!")
}
func (tm *TxManager) AddContractRawTransaction(batch *providers.BatchRequest, from dto.Address, to *dto.Address,
	amount *big.Int, gasLimit uint64, nonce uint64,
	contract string, funcName string, args ...interface{}) error {
	address, data, err := tm.Contracts.Pack(contract, funcName, args...)
	if err != nil {
		return err
	}

	if to != nil {
		address = *to
	}
	raw, err := tm.Keystore.CreateRawTx(nonce, from, &address, amount, tm.GasPrice, gasLimit, data)
	if err != nil {
		return err
	}
	tm.AddRawTransaction(batch, raw)
	return nil
}
func (tm *TxManager) UnpackResult(response *dto.RequestResult, contract string, funcName string, result interface{}) error {
	if response.Error != nil {
		return errors.New(response.Error.Message)
	}
	if response.Result == nil {
		return errors.New("Empty response")
	}

	if curContract, exist := tm.Contracts.Contracts[contract]; exist {
		return curContract.Unpack(result, funcName, common.FromHex(string(response.Result[1:len(response.Result)-1])))
	}
	return errors.New("contract is not exist")
}
