package common

import (
	"github.com/nuvla/api-client-go/types"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createMockMachineId(mId string) string {
	mockDir := "/tmp/"
	dir, err := os.MkdirTemp(mockDir, "creds")
	if err != nil {
		panic(err)
	}
	tmpDir := filepath.Join(dir, "/etc/")
	err = os.Mkdir(tmpDir, 0755)
	if !os.IsExist(err) && err != nil {
		panic(err)
	}

	if mId == "" {
		mId = "machine-id"
	}

	err = os.WriteFile(filepath.Join(dir, MachineIdFile), []byte(mId), 0644)
	if err != nil {
		panic(err)
	}

	return dir
}

func Test_EncryptCredentials(t *testing.T) {
	creds := types.ApiKeyLogInParams{
		Key:    "key",
		Secret: "secret",
	}
	dir := createMockMachineId("")
	defer os.RemoveAll(dir)

	e, err := GetIrs(creds, dir, "nuvla-edge-id")
	assert.Nil(t, err)
	assert.NotEqualf(t, "", e, "encrypted credentials should not be empty")
}

func Test_Encrypt(t *testing.T) {

	_, err := getIrs([]byte("key"), []byte("text"))
	assert.NotNil(t, err)

	_, err = getIrs([]byte(""), []byte(""))
	assert.NotNil(t, err)

	creds := types.ApiKeyLogInParams{
		Key:    "key",
		Secret: "secret",
	}

	text := creds.Key + ":" + creds.Secret
	bText := addPadding(text)
	k := strings.Repeat("K", 32)
	b, err := getIrs([]byte(k), bText)

	assert.Nil(t, err)
	assert.NotNil(t, b)
	assert.Equal(t, 32, len(b))
}

func Test_DecryptCredentials(t *testing.T) {

	mId := ""
	neId := "0e9c180e-f4a8-488a-89d4-e6ee6496b4d7"

	d := createMockMachineId(mId)
	defer os.RemoveAll(d)

	creds := types.ApiKeyLogInParams{
		Key:    "key",
		Secret: "secret",
	}

	b, err := GetIrs(creds, d, neId)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := FromIrs(b, d, neId)
	assert.Nil(t, err)
	assert.Equal(t, creds.Key, decrypted.Key)
	assert.Equal(t, creds.Secret, decrypted.Secret)
}

func Test_AddPadding(t *testing.T) {

	creds := types.ApiKeyLogInParams{
		Key:    "key",
		Secret: "secret",
	}

	credsStr := creds.Key + ":" + creds.Secret
	padded := addPadding(credsStr)

	assert.Equal(t, 16, len(padded))
	assert.Equal(t, padded[len(padded)-1], byte(16-len(credsStr)))
}

func Test_RemovePadding(t *testing.T) {

	creds := types.ApiKeyLogInParams{
		Key:    "key",
		Secret: "secret",
	}

	credsStr := creds.Key + ":" + creds.Secret
	padded := addPadding(credsStr)

	_, err := removePadding(make([]byte, 0))
	assert.NotNil(t, err)

	unpadded, err := removePadding(padded)
	assert.Nil(t, err)
	assert.Equal(t, credsStr, string(unpadded))
}

func Test_HashMachineId(t *testing.T) {
	mId := "machine-id"
	neId := "nuvla-edge-id"

	hash := hashMachineId(mId, neId)
	assert.Equal(t, 32, len(hash))
}

func Test_FindMachineID(t *testing.T) {
	mockDir := "/tmp/creds/"
	dir, err := os.MkdirTemp(mockDir, "creds")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	err = os.WriteFile(filepath.Join(mockDir, MachineIdFile), []byte("machine-id"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	mID := findMachineId("/tmp/creds/mock1")
	assert.Equal(t, "", mID)

	mId := findMachineId(mockDir)
	assert.Equal(t, "machine-id", mId)
}

func Test_GetNuvlaEdgeUuid(t *testing.T) {
	neId := "nuvlabox/nuvla-edge-id"
	neUuid := getNuvlaEdgeUuid(neId)
	assert.Equal(t, "nuvla-edge-id", neUuid)

	neId = "nuvla-edge-id"
	assert.Equal(t, "nuvla-edge-id", getNuvlaEdgeUuid(neId))

	neId = "nuvlaedge/nuvla-edge-id"
	assert.Equal(t, "nuvla-edge-id", getNuvlaEdgeUuid(neId))
}
