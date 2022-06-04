package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

type Encryptor interface {
	EncryptData(data []byte) (string, error)
	DecryptData(data string) ([]byte, error)
	GenerateRandom(size int) ([]byte, error)
	EncodeData(data string) string
}

type encryptor struct {
	key []byte
}

func New() Encryptor {
	e := &encryptor{
		key: []byte("passphrasewhichneedstobe32bytes!"),
	}
	return e
}

func (e *encryptor) EncryptData(data []byte) (string, error) {
	aesblock, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce, err := e.GenerateRandom(aesgcm.NonceSize())
	if err != nil {
		return "", err
	}

	encrypted := aesgcm.Seal(nil, nonce, data, nil)
	encrypted = append(encrypted, nonce...)
	encToString := hex.EncodeToString(encrypted)

	return encToString, nil
}

func (e *encryptor) DecryptData(data string) ([]byte, error) {
	aesblock, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	encData, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}

	nonce := encData[len(encData)-aesgcm.NonceSize():]
	enc := encData[:len(encData)-aesgcm.NonceSize()]

	decrypted, err := aesgcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func (e *encryptor) GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (e *encryptor) EncodeData(data string) string {
	buf := bytes.Buffer{}
	encoder := base64.NewEncoder(base64.URLEncoding, &buf)
	encoder.Write([]byte(data))
	return buf.String()
}
