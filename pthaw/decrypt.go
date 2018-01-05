package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/tarndt/errs"
	"github.com/tarndt/pmigrate/lib/transpenc"
)

func getSrcDecyptor(srcRdr io.Reader, encParams transpenc.EncryptionParams, keyDir string) (io.Reader, error) {
	var cipherConstructor func(cipher.Block, []byte) cipher.Stream
	switch encParams.EncryptAlgo {
	case "none", "":
		return srcRdr, nil
	case "AES-CFB":
		cipherConstructor = cipher.NewCFBDecrypter
	case "AES-CTR":
		cipherConstructor = cipher.NewCTR
	case "AES-OFB":
		cipherConstructor = cipher.NewOFB
	default:
		return nil, errs.New("Unknown encyption algorithm: %q", encParams.EncryptAlgo)
	}

	key, err := ioutil.ReadFile(filepath.Join(keyDir, encParams.KeyName))
	if err != nil {
		return nil, errs.Append(err, "Could not read decryption key file")
	}

	aesEnc, err := aes.NewCipher(key)
	if err != nil {
		return nil, errs.Append(err, "Could not create AES decryptor")
	}

	initVect, err := hex.DecodeString(encParams.InitVector)
	if err != nil {
		return nil, errs.Append(err, "Provided AES initialization vector could not bed hex decoded")
	} else if len(initVect) != aes.BlockSize {
		return nil, errs.New("Provided AES initialization vector has: %d bytes, rather than the required: %d bytes.", len(initVect), aes.BlockSize)
	}

	return cipher.StreamReader{
		S: cipherConstructor(aesEnc, initVect),
		R: srcRdr,
	}, nil
}
