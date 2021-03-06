/*
 * @Copyright Reserved By Janusec (https://www.janusec.com/).
 * @Author: U2
 * @Date: 2018-07-14 16:25:15
 * @Last Modified: U2, 2018-07-14 16:25:15
 */

package data

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"janusec/models"
	"janusec/utils"
)

var (
	// RootKey : root key for encryption
	RootKey, _  = hex.DecodeString("58309a83b94a93313a8de8f3ca815f709f4ea52066417b2ae592f2dbfd1c69ab")
	instanceKey []byte
	// NodesKey for replica nodes
	NodesKey []byte
	// HexEncryptedNodesKey for replica nodes
	HexEncryptedNodesKey string
)

// LoadInstanceKey ...
func (dal *MyDAL) LoadInstanceKey() {
	if !dal.ExistsSetting("instance_key") {
		instanceKey = GenRandomAES256Key()
		encryptedInstanceKey := AES256Encrypt(instanceKey, true)
		hexInstanceKey := hex.EncodeToString(encryptedInstanceKey)
		err := dal.SaveStringSetting("instance_key", hexInstanceKey)
		if err != nil {
			utils.DebugPrintln("LoadInstanceKey SaveStringSetting", err)
		}
	} else {
		hexEncryptedKey := dal.SelectStringSetting("instance_key")
		decodeEncryptedKey, _ := hex.DecodeString(hexEncryptedKey)
		instanceKey, _ = AES256Decrypt(decodeEncryptedKey, true)
	}
}

// LoadNodesKey ...
func (dal *MyDAL) LoadNodesKey() {
	if !dal.ExistsSetting("nodes_key") {
		NodesKey = GenRandomAES256Key()
		encryptedNodesKey := AES256Encrypt(NodesKey, true)
		HexEncryptedNodesKey = hex.EncodeToString(encryptedNodesKey)
		err := dal.SaveStringSetting("nodes_key", HexEncryptedNodesKey)
		if err != nil {
			utils.DebugPrintln("LoadNodesKey SaveStringSetting", err)
		}
	} else {
		var err error
		HexEncryptedNodesKey = dal.SelectStringSetting("nodes_key")
		decodeEncryptedKey, _ := hex.DecodeString(HexEncryptedNodesKey)
		NodesKey, err = AES256Decrypt(decodeEncryptedKey, true)
		if err != nil {
			utils.DebugPrintln("LoadNodesKey AES256Decrypt", err)
		}
	}
}

// GetHexEncryptedNodesKey ...
func GetHexEncryptedNodesKey() *models.NodesKey {
	nodesKey := &models.NodesKey{HexEncryptedKey: HexEncryptedNodesKey}
	return nodesKey
}

// GenRandomAES256Key ...
func GenRandomAES256Key() []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		utils.DebugPrintln("GenRandomAES256Key", err)
	}
	return key
}

// EncryptWithKey ...
func EncryptWithKey(plaintext []byte, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		utils.DebugPrintln("EncryptWithKey NewCipher", err)
	}
	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		utils.DebugPrintln("EncryptWithKey ReadFull", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		utils.DebugPrintln("EncryptWithKey NewGCM", err)
	}
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext
}

// AES256Encrypt ...
func AES256Encrypt(plaintext []byte, useRootkey bool) []byte {
	key := instanceKey
	if useRootkey {
		key = RootKey
	}
	ciphertext := EncryptWithKey(plaintext, key)
	return ciphertext
}

// DecryptWithKey ...
func DecryptWithKey(ciphertext []byte, key []byte) ([]byte, error) {
	var block cipher.Block
	var err error
	block, err = aes.NewCipher(key)
	if err != nil {
		utils.DebugPrintln("DecryptWithKey NewCipher", err)
		return []byte{}, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		utils.DebugPrintln("DecryptWithKey NewGCM", err)
		return []byte{}, err
	}
	nonce, ciphertext := ciphertext[:12], ciphertext[12:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		utils.DebugPrintln("DecryptWithKey Open", err)
		return []byte{}, err
	}
	return plaintext, nil
}

// AES256Decrypt ...
func AES256Decrypt(ciphertext []byte, useRootkey bool) ([]byte, error) {
	key := instanceKey
	if useRootkey {
		key = RootKey
	}
	plaintext, err := DecryptWithKey(ciphertext, key)
	if err != nil {
		utils.DebugPrintln("AES256Decrypt", err)
	}
	return plaintext, err
}

// GetRandomSaltString ...
func GetRandomSaltString() string {
	salt := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		utils.DebugPrintln("GetRandomSaltString", err)
	}
	saltStr := fmt.Sprintf("%x", salt)
	return saltStr
}

// SHA256Hash ...
func SHA256Hash(plaintext string) string {
	hash := sha256.New()
	_, err := hash.Write([]byte(plaintext))
	if err != nil {
		utils.DebugPrintln("SHA256Hash hash.Write", err)
	}
	result := fmt.Sprintf("%x", hash.Sum(nil))
	return result
}

// NodeHexKeyToCryptKey ...
func NodeHexKeyToCryptKey(hexKey string) []byte {
	encrptedKey, err := hex.DecodeString(hexKey)
	if err != nil {
		utils.DebugPrintln("NodeHexKeyToCryptKey DecodeString", err)
	}
	key, err := AES256Decrypt(encrptedKey, true)
	if err != nil {
		utils.DebugPrintln("NodeHexKeyToCryptKey AES256Decrypt", err)
	}
	return key
}

// CryptKeyToNodeHexKey ...
func CryptKeyToNodeHexKey(keyBytes []byte) string {
	encryptedKey := AES256Encrypt(keyBytes, true)
	hexKey := hex.EncodeToString(encryptedKey)
	return hexKey
}
