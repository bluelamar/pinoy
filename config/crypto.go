package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

// decrypt from base64 to decrypted string
func Decrypt(keyText, cryptoText string) string {

	// key := []byte(keyText)
	key := normalizeKey(keyText)
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}

// encrypt string then encode to base64
func Encrypt(keyText, clearText string) string {

	//key := []byte(keyText)
	key := normalizeKey(keyText)
	log.Println("FIX encrypt key len=", len(key))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	text := []byte(clearText)
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(text))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], text)
	cryptoText := base64.URLEncoding.EncodeToString(cipherText)
	return cryptoText
}

func normalizeKey(keyText string) []byte {
	// ensure key will be of correct length for AES: 128, 192 or 256 bit key length. Which is 16, 24 or 32 bytes
	klen := len(keyText)
	paddingLen := 0
	if klen < 16 {
		paddingLen = 16 - klen
	} else if klen < 24 {
		paddingLen = 24 - klen
	} else if klen < 32 {
		paddingLen = 32 - klen
	} else if klen > 32 {
		// truncate the key
		paddingLen = -32
	}

	var key []byte
	if paddingLen >= 0 {
		key = make([]byte, klen+paddingLen)
	} else {
		key = make([]byte, 32)
	}

	if paddingLen != 0 {
		log.Println("encrypt: key is either too short or too long: len=", klen, " so must truncate to 32 or pad by ", paddingLen, " bytes")
	}
	for i, v := range []byte(keyText) {
		if i == 32 {
			break
		}
		key[i] = v
	}
	for i := 0; i < paddingLen; i++ {
		key[i+klen] = byte(('a' + i) % 26)
	}
	return key
}

func HashIt(text string) string {
	h := md5.New()
	sum := h.Sum([]byte(text))
	cryptoText := base64.URLEncoding.EncodeToString(sum)
	return cryptoText
}
