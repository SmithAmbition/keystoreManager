package txManager

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-web3/dto"
	"github.com/ethereum/go-web3/web3Manager"
	"github.com/ethereum/keystoreManager/contract"
)

type KeystoreInterface interface {
	CreateRawTx(nonce uint64, from dto.Address, to *dto.Address,
		amount *big.Int, gasPrice *big.Int, gasLimit uint64,
		data []byte) ([]byte, error)
	Unlock(addr dto.Address, password string) error
	NewAccount(password string) (dto.Address, error)
	GetDecryptedKey(a accounts.Account, auth string) (accounts.Account, *keystore.Key, error)
}
type TxManager struct {
	Keystore  KeystoreInterface
	Contracts *contract.ContractManager
	Web3      *web3Manager.Web3Manager
	GasPrice  *big.Int
}

func (tm *TxManager) getGasPrice() (*big.Int, error) {
	return tm.GasPrice, nil
	gasPrice, err := tm.Web3.Eth.GetGasPrice()
	if err != nil {
		return nil, err
	}
	addValue := new(big.Int).Div(gasPrice, big.NewInt(2))
	if addValue.Cmp(big.NewInt(50e9)) > 0 {
		addValue = big.NewInt(50e9)
	}
	gasPrice = gasPrice.Add(gasPrice, addValue)
	return gasPrice, nil
}
func (tm *TxManager) DeployContractRawTransaction(from dto.Address,
	amount *big.Int, gasLimit uint64,
	contract string, bytecode []byte, args ...interface{}) (common.Address, string, error) {
	data, err := tm.Contracts.DeployPack(contract, bytecode, args...)
	if err != nil {
		return common.Address{}, "", err
	}
	nonce, err := tm.Web3.Eth.GetTransactionCount(from.String(), "latest")
	if err != nil {
		return common.Address{}, "", err
	}
	gasPrice, err := tm.getGasPrice()
	if err != nil {
		return common.Address{}, "", err
	}
	raw, err := tm.Keystore.CreateRawTx(nonce.Uint64(), from, nil, amount, gasPrice, gasLimit, data)
	if err != nil {
		return common.Address{}, "", err
	}
	address := crypto.CreateAddress(common.Address(from), nonce.Uint64())
	hash, err := tm.Web3.Eth.SendRawTransaction(raw)
	return address, hash, err
}
func (tm *TxManager) SendRawTransaction(from dto.Address, to *dto.Address,
	amount *big.Int, gasLimit uint64, data []byte) (string, error) {
	nonce, err := tm.Web3.Eth.GetTransactionCount(from.String(), "latest")
	if err != nil {
		return "", err
	}
	gasPrice, err := tm.getGasPrice()
	if err != nil {
		return "", err
	}
	raw, err := tm.Keystore.CreateRawTx(nonce.Uint64(), from, to, amount, gasPrice, gasLimit, data)
	if err != nil {
		return "", err
	}
	return tm.Web3.Eth.SendRawTransaction(raw)
}
func (tm *TxManager) SendContractRawTransaction(from dto.Address, to *dto.Address,
	amount *big.Int, gasLimit uint64,
	contract string, funcName string, args ...interface{}) (string, error) {
	nonce, err := tm.Web3.Eth.GetTransactionCount(from.String(), "latest")
	fmt.Println("nonce: ", nonce)
	if err != nil {
		return "", err
	}
	return tm.SendContractRawTransactionWithNonce(from, to, amount, gasLimit, nonce.Uint64(), contract, funcName, args...)
}
func (tm *TxManager) SendContractRawTransactionWithNonce(from dto.Address, to *dto.Address,
	amount *big.Int, gasLimit uint64, nonce uint64,
	contract string, funcName string, args ...interface{}) (string, error) {
	address, data, err := tm.Contracts.Pack(contract, funcName, args...)
	if err != nil {
		return "", err
	}
	if to != nil {
		address = *to
	}
	gasPrice, err := tm.getGasPrice()
	if err != nil {
		return "", err
	}
	raw, err := tm.Keystore.CreateRawTx(nonce, from, &address, amount, gasPrice, gasLimit, data)
	if err != nil {
		return "", err
	}
	return tm.Web3.Eth.SendRawTransaction(raw)
}
func (tm *TxManager) SendContractCall(from dto.Address, to *dto.Address, blockNumber *big.Int, result interface{}, contract string, funcName string, args ...interface{}) error {
	gasPrice, err := tm.getGasPrice()
	if err != nil {
		return err
	}
	if curContract, exist := tm.Contracts.Contracts[contract]; exist {
		data, err := curContract.Pack(funcName, args...)
		address := common.Address(curContract.Address)
		if to != nil {
			address = common.Address(*to)
		}
		msg := &ethereum.CallMsg{
			From:     common.Address(from),
			To:       &address,
			Gas:      6750000,
			GasPrice: gasPrice,
			Data:     data,
		}
		response, err := tm.Web3.Eth.DoCall(msg, blockNumber)
		if err != nil {
			return err
		}
		return curContract.Unpack(result, funcName, common.FromHex(string(response.Result[1:len(response.Result)-1])))
	}
	return errors.New("contract is not exist")
}
