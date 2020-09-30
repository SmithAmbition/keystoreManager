package defaultKeystore

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-web3/dto"
	"golang.org/x/crypto/sha3"
)

const (
	decryptExtra = "func (ks *DefaultKeystore) getDecryptedKey(a accounts.Account, auth string) (accounts.Account, *keystore.Key, error)"
)

type DefaultKeystore struct {
	Keystore *keystore.KeyStore
	ChainID  *big.Int
}

func NewKeystore(keystoreDir string, chainId int64) *DefaultKeystore {
	ks := &DefaultKeystore{}
	ks.ChainID = big.NewInt(chainId)
	ks.Keystore = keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	return ks
}
func (ks *DefaultKeystore) GetDecryptedKey(a accounts.Account, auth string) (accounts.Account, *keystore.Key, error) {
	a, err := ks.Keystore.Find(a)
	if err != nil {
		return a, nil, err
	}
	key, err := ks.GetKey(a.Address, a.URL.Path, auth)
	return a, key, err
}
func (ks *DefaultKeystore) Unlock(addr dto.Address, password string) error {
	return ks.TimedUnlock(addr, password, 0)
}
func (ks *DefaultKeystore) TimedUnlock(addr dto.Address, password string, timeout time.Duration) error {
	decryptPass := ks.DecryptPassphrase(password)
	acc := accounts.Account{Address: common.Address(addr)}
	return ks.Keystore.TimedUnlock(acc, decryptPass, timeout)
}
func (ks *DefaultKeystore) GetKey(addr common.Address, filename, auth string) (*keystore.Key, error) {
	decryptPass := ks.DecryptPassphrase(auth)
	// Load the key from the keystore and decrypt its contents
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keyjson, decryptPass)
	if err != nil {
		return nil, err
	}
	// Make sure we're really operating on the requested key (no swap attacks)
	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
	}
	return key, nil
}

func (ks *DefaultKeystore) DecryptPassphrase(passphrase string) string {
	hash := RlpHash([]string{decryptExtra, passphrase})
	return common.Bytes2Hex(hash[:10])
}
func (ks *DefaultKeystore) NewAccount(passphrase string) (dto.Address, error) {
	decryptPass := ks.DecryptPassphrase(passphrase)
	account, err := ks.Keystore.NewAccount(decryptPass)
	return dto.Address(account.Address), err
}
func (ks *DefaultKeystore) SignTxHash(hash []byte, from dto.Address) (sig []byte, err error) {
	prv, err := ks.Keystore.GetUnlockedKey(common.Address(from))
	if err != nil {
		return nil, err
	}
	return crypto.Sign(hash, prv)
}
func RlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// SignatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (ks *DefaultKeystore) SignatureValues(sig []byte) (r, s, v *big.Int, err error) {
	if len(sig) != crypto.SignatureLength {
		panic(fmt.Sprintf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	if ks.ChainID.Sign() != 0 {
		chainIdMul := new(big.Int).Mul(ks.ChainID, big.NewInt(2))
		v = big.NewInt(int64(sig[64] + 35))
		v.Add(v, chainIdMul)
	}
	return r, s, v, nil
}
