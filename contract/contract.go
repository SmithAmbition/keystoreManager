package contract

import (
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-web3/dto"
)

type Contract struct {
	Abi     abi.ABI
	Address dto.Address
}

func (contract *Contract) Pack(name string, args ...interface{}) ([]byte, error) {
	return contract.Abi.Pack(name, args...)
}
func (contract *Contract) Unpack(v interface{}, name string, data []byte) (err error) {
	return contract.Abi.Unpack(v, name, data)
}
func (contract *Contract) ReadAbi(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	contract.Abi, err = abi.JSON(file)
	return err
}
func (contract *Contract) Deploy(bytecode []byte, args ...interface{}) ([]byte, error) {
	input, err := contract.Abi.Pack("", args...)
	if err != nil {
		return nil, err
	}
	return append(bytecode, input...), nil
}
