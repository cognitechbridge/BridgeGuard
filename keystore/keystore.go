package keystore

import (
	"crypto/rand"
	"ctb-cli/crypto/key_crypto"
	"ctb-cli/crypto/recovery"
	"ctb-cli/filesyetem/key_repository"
	"ctb-cli/types"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"io"
	"os"
)

type KeyStorer interface {
	Get(keyID string) (*types.Key, error)
	Insert(keyID string, key types.Key) error
	GetRecoveryItems() ([]types.RecoveryItem, error)
	AddRecoveryKey(inPath string) error
	GenerateClientKeys() (err error)
}

type Key = types.Key

// KeyStoreDefault represents a key store
type KeyStoreDefault struct {
	clintId       string
	rootKey       Key
	privateKey    []byte
	recoveryItems []types.RecoveryItem
	keyRepository key_repository.KeyRepository
}

var _ KeyStorer = &KeyStoreDefault{}

// NewKeyStore creates a new instance of KeyStoreDefault
func NewKeyStore(clientId string, rootKey Key, keyRepository key_repository.KeyRepository) *KeyStoreDefault {
	return &KeyStoreDefault{
		clintId:       clientId,
		rootKey:       rootKey,
		keyRepository: keyRepository,
		recoveryItems: make([]types.RecoveryItem, 0),
	}
}

// Insert inserts a key into the key store
func (ks *KeyStoreDefault) Insert(keyID string, key Key) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}

	pk, err := ks.getPublicKey()
	if err != nil {
		return err
	}
	keyHashed, err := key_crypto.SealDataKey(key[:], pk)
	if err != nil {
		return err
	}

	return ks.keyRepository.SaveDataKey(keyID, keyHashed, ks.clintId)
}

// Get retrieves a key from the key store
func (ks *KeyStoreDefault) Get(keyID string) (*Key, error) {
	if err := ks.LoadKeys(); err != nil {
		return nil, fmt.Errorf("cannot load keys: %v", err)
	}
	sk, err := ks.keyRepository.GetDataKey(keyID)
	if err != nil {
		return nil, err
	}
	key, err := key_crypto.OpenDataKey(sk, ks.privateKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (ks *KeyStoreDefault) Share(keyId string, recipient []byte, recipientClientId string) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}

	key, err := ks.Get(keyId)
	if err != nil {
		return fmt.Errorf("cannot load key: %v", err)
	}

	keyHashed, err := key_crypto.SealDataKey(key[:], recipient)
	if err != nil {
		return err
	}

	return ks.keyRepository.SaveDataKey(keyId, keyHashed, recipientClientId)
}

func (ks *KeyStoreDefault) GetRecoveryItems() ([]types.RecoveryItem, error) {
	return ks.recoveryItems, nil
}

func (ks *KeyStoreDefault) AddRecoveryKey(inPath string) error {
	path := os.ExpandEnv(inPath)
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	rec, err := recovery.UnmarshalRecoveryItem(pemBytes)
	if err != nil {
		return err
	}

	ks.recoveryItems = append(ks.recoveryItems, *rec)

	return nil
}

func (ks *KeyStoreDefault) LoadKeys() error {

	if ks.privateKey != nil {
		return nil
	}
	serializedPrivateKey, err := ks.keyRepository.GetPrivateKey()
	if err != nil {
		return err
	}
	ks.privateKey, err = key_crypto.OpenPrivateKey(serializedPrivateKey, &ks.rootKey)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStoreDefault) GenerateClientKeys() (err error) {
	//Generate private key
	privateKey := types.Key{}
	io.ReadFull(rand.Reader, privateKey[:])
	//Save private key
	sealPrivateKey, err := key_crypto.SealPrivateKey(privateKey[:], &ks.rootKey)
	if err != nil {
		return err
	}
	err = ks.keyRepository.SavePrivateKey(sealPrivateKey)
	if err != nil {
		return err
	}

	publicKey, err := ks.getPublicKey()
	if err != nil {
		return err
	}
	serializedPublic, err := key_crypto.SerializePublicKey(publicKey)
	if err != nil {
		return err
	}
	err = ks.keyRepository.SavePublicKey(ks.clintId, serializedPublic)
	if err != nil {
		return err
	}
	err = ks.LoadKeys()
	return
}

func (ks *KeyStoreDefault) getPublicKey() ([]byte, error) {
	return curve25519.X25519(ks.privateKey, curve25519.Basepoint)
}
