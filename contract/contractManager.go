package contract

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ethereum/go-web3/dto"
)

type ContractManager struct {
	Contracts map[string]*Contract
}

func NewContractManager(fileName string) *ContractManager {
	cm := &ContractManager{
		Contracts: make(map[string]*Contract),
	}
	cm.ReadContracts(fileName)
	return cm
}

type ContractInfo struct {
	Address  string `json:"address"`
	FileName string `json:"fileName"`
}

func (cm *ContractManager) ReadContracts(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	contractInfo := make(map[string]ContractInfo)
	if err := json.NewDecoder(file).Decode(&contractInfo); err != nil {
		return err
	}
	path := filepath.Dir(fileName)
	for key, info := range contractInfo {
		newContract := &Contract{}
		newContract.ReadAbi(filepath.Join(path, info.FileName))
		newContract.Address = dto.HexToAddress(info.Address)
		cm.Contracts[key] = newContract
	}
	return nil
}
func (cm *ContractManager) SaveContracts(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	contractInfo := make(map[string]*ContractInfo)
	if err := json.NewDecoder(file).Decode(&contractInfo); err != nil {
		file.Close()
		return err
	}
	file.Close()
	path := filepath.Dir(fileName)
	for key, info := range cm.Contracts {
		contractInfo[key].Address = info.Address.String()
		out, _ := json.MarshalIndent(info.Abi, "", "  ")
		if err := ioutil.WriteFile(filepath.Join(path, contractInfo[key].FileName), out, 0644); err != nil {
			return err
		}
	}
	out, _ := json.MarshalIndent(contractInfo, "", "  ")
	if err := ioutil.WriteFile(fileName, out, 0644); err != nil {
		return err
	}
	return nil
}
func (cm *ContractManager) readContractAbi(contractName string, abiFile string) error {
	newContract := &Contract{}
	err := newContract.ReadAbi(abiFile)
	if err != nil {
		return err
	}
	cm.Contracts[contractName] = newContract
	return nil
}
func (cm *ContractManager) Pack(contractName string, name string, args ...interface{}) (dto.Address, []byte, error) {
	if contract, exist := cm.Contracts[contractName]; exist {
		result, err := contract.Pack(name, args...)
		return contract.Address, result, err
	}
	return dto.Address{}, nil, errors.New("contract is not exist")
}
func (cm *ContractManager) DeployPack(contractName string, bytecode []byte, args ...interface{}) ([]byte, error) {
	if contract, exist := cm.Contracts[contractName]; exist {
		result, err := contract.Deploy(bytecode, args...)
		return result, err
	}
	return nil, errors.New("contract is not exist")
}
