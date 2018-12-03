package offlinetxs

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/matheusd/offline-tx-signer/pkg/snacl"
)

func StoreLocalSeed(path string, seed, passphrase []byte) error {
	secretKey, err := snacl.NewSecretKey(&passphrase, snacl.DefaultN,
		snacl.DefaultR, snacl.DefaultP)
	if err != nil {
		return err
	}
	defer secretKey.Zero()

	keyData := secretKey.Marshal()
	if len(keyData) != 88 {
		return fmt.Errorf("marshal() didn't return the correct len")
	}

	encryptedSeed, err := secretKey.Encrypt(seed)
	if err != nil {
		return err
	}

	empty := [64]byte{}
	fullData := empty[:]                    // 64 byte empty prefix
	fullData = append(fullData, keyData...) // 88 byte marshalled key data
	fullData = append(fullData, encryptedSeed...)

	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	f, err := os.OpenFile(path, flags, os.FileMode(0600))
	if err != nil {
		return err
	}

	n, err := f.Write(fullData)
	if (err != nil) || (n != len(fullData)) {
		f.Close() //ignore error here, due to having already errored in writing
		if err != nil {
			return err
		}
		return fmt.Errorf("Could not write all data bytes")
	}

	return f.Close()
}

func RetrieveLocalSeed(path string, passphrase []byte) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fullData := make([]byte, 256) // 256 should be enough.
	n, err := f.Read(fullData)
	f.Close()
	if err != nil && (err != io.EOF) {
		return nil, err
	}

	if n < 184 {
		return nil, fmt.Errorf("Read less than the bare minimum needed")
	}

	var empty [64]byte
	if !bytes.Equal(fullData[:64], empty[:]) {
		return nil, fmt.Errorf("Prefix mismatch in file")
	}

	secretKey := snacl.SecretKey{}
	defer secretKey.Zero()

	err = secretKey.Unmarshal(fullData[64:152])
	if err != nil {
		return nil, err
	}

	err = secretKey.DeriveKey(&passphrase)
	if err != nil {
		return nil, err
	}

	seed, err := secretKey.Decrypt(fullData[152:n])
	if err != nil {
		return nil, err
	}

	return seed, nil
}
