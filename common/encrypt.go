package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/nuvla/api-client-go/common"
	"github.com/nuvla/api-client-go/types"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const MachineIdFile = "/etc/machine-id"

func GetIrs(creds types.ApiKeyLogInParams, rootFsPath string, nuvlaEdgeId string) (string, error) {
	key := buildKey(rootFsPath, nuvlaEdgeId)

	plainText := addPadding(creds.Key + ":" + creds.Secret)

	irs, err := getIrs(key, plainText)
	if err != nil {
		log.Errorf("Error conforming IRS: %s", err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(irs), nil
}

func FromIrs(irs64 string, rootFsPath string, nuvlaEdgeId string) (types.ApiKeyLogInParams, error) {
	irs, err := base64.StdEncoding.DecodeString(irs64)
	if err != nil {
		return types.ApiKeyLogInParams{}, fmt.Errorf("error decoding IRS 64: %s", err)
	}

	key := buildKey(rootFsPath, nuvlaEdgeId)

	dIrs, err := fromIrs(key, irs)
	if err != nil {
		return types.ApiKeyLogInParams{}, fmt.Errorf("error decrypting credentials: %s", err)
	}

	dIrs, err = removePadding(dIrs)
	if err != nil {
		return types.ApiKeyLogInParams{}, fmt.Errorf("error removing padding: %s", err)
	}

	lSlice := strings.Split(string(dIrs), ":")
	if len(lSlice) != 2 {
		return types.ApiKeyLogInParams{}, errors.New("invalid credentials")
	}

	return types.ApiKeyLogInParams{Key: lSlice[0], Secret: lSlice[1]}, nil
}

func buildKey(rootFs, neId string) []byte {
	return hashMachineId(findMachineId(rootFs), getNuvlaEdgeUuid(neId))
}

func addPadding(toEncrypt string) []byte {
	padding := aes.BlockSize - len(toEncrypt)%aes.BlockSize
	padtext := strings.Repeat(string(byte(padding)), padding)
	return []byte(toEncrypt + padtext)
}

func removePadding(plainText []byte) ([]byte, error) {
	if len(plainText) == 0 {
		return nil, errors.New("plainText data is empty")
	}

	paddingLen := int(plainText[len(plainText)-1])

	if paddingLen > len(plainText) {
		return nil, errors.New("invalid padding length")
	}

	return plainText[:len(plainText)-paddingLen], nil
}

func fromIrs(key, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %s", err)
	}

	if len(cipherText) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	if len(cipherText)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	cipher.NewCBCDecrypter(block, iv).CryptBlocks(cipherText, cipherText)

	return cipherText, nil
}

func getIrs(key, plainText []byte) ([]byte, error) {
	if len(plainText)%aes.BlockSize != 0 {
		return nil, errors.New("text to getIrs len is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %s", err)
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("error generating random IV: %s", err)
	}

	cipher.NewCBCEncrypter(block, iv).CryptBlocks(cipherText[aes.BlockSize:], plainText)

	return cipherText, nil
}

func hashMachineId(machineId, neId string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(neId + ":" + machineId))
	return hasher.Sum(nil)
}

func findMachineId(rootFsPath string) string {
	mIdFile := MachineIdFile
	if rootFsPath != "" {
		mIdFile = filepath.Join(rootFsPath, mIdFile)
	}

	if !common.FileExists(mIdFile) {
		log.Errorf("Machine-id file %s does not exist", mIdFile)
		return ""
	}

	f, err := os.ReadFile(mIdFile)
	if err != nil {
		log.Errorf("Error reading machine-id file %s: %s", mIdFile, err)
		return ""
	}
	return strings.Trim(string(f), "\n\r \t")
}

func getNuvlaEdgeUuid(nuvlaEdgeId string) string {
	if strings.HasPrefix(nuvlaEdgeId, "nuvlabox/") {
		return strings.TrimPrefix(nuvlaEdgeId, "nuvlabox/")
	}

	if strings.HasPrefix(nuvlaEdgeId, "nuvlaedge/") {
		return strings.TrimPrefix(nuvlaEdgeId, "nuvlaedge/")
	}

	return nuvlaEdgeId
}
