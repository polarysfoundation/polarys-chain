package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	pec256 "github.com/polarysfoundation/pec-256"
	pm256 "github.com/polarysfoundation/pm-256"
	"github.com/polarysfoundation/polarys-chain/modules/common"
	"github.com/polarysfoundation/polarys-chain/modules/crypto"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

var (
	keystoreDir = getKeystorePath()
)

var (
	logger = logrus.New()
)

const (
	basePath     = ".polarys"
	keystorePath = "keystore"
)

func getKeystorePath() string {
	return filepath.Join(getDir(basePath, keystorePath))
}

func getDir(dir string, subdir string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}
	subDirPath := filepath.Join(homeDir, dir, subdir)

	err = os.MkdirAll(subDirPath, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create %s directory: %v", subdir, err)
	}

	return subDirPath
}

type Keystore struct {
	Address common.Address `json:"address"`
	Crypto  Crypto         `json:"crypto"`
	Version int            `json:"version"`
}

type Crypto struct {
	Ciphertext   string `json:"ciphertext"`
	Cipherparams struct {
		IV string `json:"iv"`
	} `json:"cipherparams"`
	Cipher    string `json:"cipher"`
	KDF       string `json:"kdf"`
	KDFParams struct {
		DKLen int    `json:"dklen"`
		Salt  string `json:"salt"`
		N     int    `json:"n"`
		R     int    `json:"r"`
		P     int    `json:"p"`
	} `json:"kdfparams"`
	MAC string `json:"mac"`
}

func NewKeypair(passphrase []byte) (*Keypair, error) {
	priv, pub := crypto.GenerateKey()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	keypair := &Keypair{
		priv: priv,
		pub:  pub,
	}

	err := saveKeystore(*keypair, passphrase)
	if err != nil {
		logger.Errorf("error saving keystore: %v", err)
		return nil, err
	}

	return keypair, nil
}

func GetLocalAccounts() []common.Address {
	addresses := make([]common.Address, 0)

	ok, keysLen := HasKeyspair()
	if !ok || keysLen == 0 {
		return addresses
	}

	files, err := os.ReadDir(keystoreDir)
	if err != nil {
		return addresses
	}

	addressSet := make(map[common.Address]struct{}) // Usamos un set para evitar duplicados

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".json")
			addr := common.CXIDToAddress(nameWithoutExt)

			if _, exists := addressSet[addr]; !exists {
				addresses = append(addresses, addr)
				addressSet[addr] = struct{}{}
			}
		}
	}

	return addresses
}

func HasKeyspair() (bool, int) {
	files, err := os.ReadDir(keystoreDir)
	if err != nil {
		return false, 0
	}

	if len(files) == 0 {
		return false, 0
	}

	return true, len(files)
}

func ExistInLocal(address common.Address) bool {
	addresses := GetLocalAccounts()

	for _, addr := range addresses {
		if addr.String() == address.String() {
			return true
		}
	}

	return false
}

func LoadKeypairs(passphrase []byte) ([]*Keypair, error) {
	files, err := os.ReadDir(keystoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore directory: %v", err)
	}

	var keypairs []*Keypair
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			keypair, err := loadKeypairFromKeystore(filepath.Join(keystoreDir, file.Name()), passphrase)
			if err != nil {
				fmt.Printf("failed to load wallet from %s: %v\n", file.Name(), err)
				continue
			}
			keypairs = append(keypairs, keypair)
		}
	}

	return keypairs, nil
}

func GetKeypairByAddress(address common.Address, passphrase []byte) (*Keypair, error) {
	if !ExistInLocal(address) {
		return nil, fmt.Errorf("address %s not found", address.String())
	}

	fileName := fmt.Sprintf("%s.json", address.String())
	file := filepath.Join(keystoreDir, fileName)

	keypair, err := loadKeypairFromKeystore(file, passphrase)
	if err != nil {
		return nil, fmt.Errorf("error getting keypair for address %s", address.String())
	}

	return keypair, nil
}

func loadKeypairFromKeystore(filePath string, passphrase []byte) (*Keypair, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %v", err)
	}

	var keystore Keystore
	err = json.Unmarshal(data, &keystore)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal keystore file: %v", err)
	}

	privKey, pubKey, err := decryptPrivateKey(keystore.Crypto, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	return &Keypair{
		priv:      privKey,
		pub:       pubKey,
		timestamp: uint64(time.Now().Unix()) + 120,
	}, nil
}

func decryptPrivateKey(crypto Crypto, passphrase []byte) (pec256.PrivKey, pec256.PubKey, error) {
	salt, err := hex.DecodeString(crypto.KDFParams.Salt)
	if err != nil {
		return pec256.PrivKey{}, pec256.PubKey{}, fmt.Errorf("failed to decode salt: %v", err)
	}

	dk, err := scrypt.Key(passphrase, salt, crypto.KDFParams.N, crypto.KDFParams.R, crypto.KDFParams.P, crypto.KDFParams.DKLen)
	if err != nil {
		return pec256.PrivKey{}, pec256.PubKey{}, fmt.Errorf("failed to derive key: %v", err)
	}

	ciphertext, err := hex.DecodeString(crypto.Ciphertext)
	if err != nil {
		return pec256.PrivKey{}, pec256.PubKey{}, fmt.Errorf("failed to decode ciphertext: %v", err)
	}

	iv, err := hex.DecodeString(crypto.Cipherparams.IV)
	if err != nil {
		return pec256.PrivKey{}, pec256.PubKey{}, fmt.Errorf("failed to decode IV: %v", err)
	}

	block, err := aes.NewCipher(dk[:16])
	if err != nil {
		return pec256.PrivKey{}, pec256.PubKey{}, fmt.Errorf("failed to create cipher: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return pec256.PrivKey{}, pec256.PubKey{}, fmt.Errorf("failed to create GCM: %v", err)
	}

	plaintext, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return pec256.PrivKey{}, pec256.PubKey{}, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	privKey := make([]byte, 32)
	pubKey := make([]byte, 32)

	privKey = append(privKey, plaintext[:32]...)
	pubKey = append(pubKey, plaintext[32:]...)

	return pec256.BytesToPrivKey(privKey), pec256.BytesToPubKey(pubKey), nil
}

func encryptPrivateKey(privKey pec256.PrivKey, pubKey pec256.PubKey, passphrase []byte) (string, string, string, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", "", "", fmt.Errorf("failed to generate salt: %v", err)
	}

	dk, err := scrypt.Key(passphrase, salt, 1<<18, 8, 1, 32)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to derive key: %v", err)
	}

	block, err := aes.NewCipher(dk[:16])
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create cipher: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	toEncrypt := make([]byte, 64)

	toEncrypt = append(toEncrypt[:32], privKey.Bytes()...)
	toEncrypt = append(toEncrypt[32:], pubKey.Bytes()...)

	ciphertext := aesGCM.Seal(nil, nonce, toEncrypt, nil)
	return hex.EncodeToString(ciphertext), hex.EncodeToString(nonce), hex.EncodeToString(salt), nil
}

func saveKeystore(keypair Keypair, passphrase []byte) error {
	ciphertext, nonce, salt, err := encryptPrivateKey(keypair.priv, keypair.pub, passphrase)
	if err != nil {
		return err
	}

	addr := crypto.PubKeyToAddress(keypair.pub)

	keystore := Keystore{
		Address: addr,
		Version: 3,
	}
	keystore.Crypto.Ciphertext = ciphertext
	keystore.Crypto.Cipherparams.IV = nonce
	keystore.Crypto.Cipher = "aes-128-ctr"
	keystore.Crypto.KDF = "scrypt"
	keystore.Crypto.KDFParams.DKLen = 32
	keystore.Crypto.KDFParams.Salt = salt
	keystore.Crypto.KDFParams.N = 1 << 18
	keystore.Crypto.KDFParams.R = 8
	keystore.Crypto.KDFParams.P = 1
	mac := pm256.Sum256([]byte(ciphertext))
	keystore.Crypto.MAC = hex.EncodeToString(mac[:])

	data, err := json.MarshalIndent(keystore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %v", err)
	}

	if _, err := os.Stat(keystoreDir); os.IsNotExist(err) {
		if err := os.Mkdir(keystoreDir, 0700); err != nil {
			return fmt.Errorf("failed to create keystore directory: %v", err)
		}
	}

	fileName := filepath.Join(keystoreDir, addr.String()+".json")
	return os.WriteFile(fileName, data, 0644)
}
