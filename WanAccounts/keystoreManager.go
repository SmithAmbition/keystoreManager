package WanAccounts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-web3/dto"
	"github.com/ethereum/keystoreManager/defaultKeystore"
)

type KeystoreManager struct {
	*defaultKeystore.DefaultKeystore
}

func NewKeystoreManager(keystoreDir string, chainId int64) *KeystoreManager {
	ks := &KeystoreManager{defaultKeystore.NewKeystore(keystoreDir, chainId)}
	return ks
}

func (ks *KeystoreManager) SignTx(sendTx *Transaction, from dto.Address) ([]byte, error) {
	h := sendTx.RlpHashEIP155(ks.ChainID)
	sig, err := ks.SignTxHash(h[:], from)
	if err != nil {
		return nil, err
	}
	r, s, v, err := ks.SignatureValues(sig)
	if err != nil {
		return nil, err
	}
	tx, err := sendTx.CopyWithRSV(r, s, v)
	if err != nil {
		return nil, err
	}
	return rlp.EncodeToBytes(tx)
}

func (ks *KeystoreManager) CreateRawTx(nonce uint64, from dto.Address, to *dto.Address,
	amount *big.Int, gasPrice *big.Int, gasLimit uint64,
	data []byte) ([]byte, error) {
	gas := new(big.Int).SetUint64(gasLimit)
	if to != nil {
		tx := NewTransaction(nonce, common.Address(*to), amount, gas, gasPrice, data)
		return ks.SignTx(tx, from)
	} else {
		tx := NewContractCreation(nonce, amount, gas, gasPrice, data)
		return ks.SignTx(tx, from)
	}

}
