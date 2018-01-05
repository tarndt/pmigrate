package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/tarndt/errs"
	"github.com/tarndt/pmigrate/lib/transpenc"
)

func getDestEncryptor(dstWtr io.Writer, encrypt string, transpEnc *transpenc.TranportEncoding) (io.WriteCloser, error) {
	if encrypt == "" || strings.ToLower(encrypt) == "none" {
		transpEnc.EncParams.EncryptAlgo = "none"
		return newNopWriteCloser(dstWtr), nil
	}

	parts := strings.Split(encrypt, ":")
	if len(parts) != 2 {
		return nil, errs.New("Encryption parameter must be in the form <ALGO>:<PATH TO KEY>.")
	}
	algo, keypath := strings.ToUpper(parts[0]), parts[1]
	if !strings.HasPrefix(algo, "AES-") ||
		!(strings.HasSuffix(algo, "-CFB") || strings.HasSuffix(algo, "-CTR") || strings.HasSuffix(algo, "-OFB")) {
		return nil, errs.New("Encyption algorithm unknown, valid options are AES-CFB, AES-CTR (recomended) & AES-OFB")
	}

	key, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, errs.Append(err, "Could not read encryption key file")
	}

	aesEnc, err := aes.NewCipher(key)
	if err != nil {
		return nil, errs.Append(err, "Could not create AES encryptor")
	}

	initVect := make([]byte, aes.BlockSize)
	if _, err = rand.Read(initVect); err != nil {
		return nil, errs.Append(err, "Could not read entropy source to populate AES initialization vector")
	}

	var cipherConstructor func(cipher.Block, []byte) cipher.Stream
	switch algo {
	case "AES-CFB":
		cipherConstructor = cipher.NewCFBEncrypter
	case "AES-CTR":
		cipherConstructor = cipher.NewCTR
	case "AES-OFB":
		cipherConstructor = cipher.NewOFB
	default: //This should not be possible due to the check above:
		return nil, errs.New("BUG: Encyption algorithm check failed for: %q", algo)
	}

	transpEnc.EncParams.KeyName = filepath.Base(keypath)
	transpEnc.EncParams.EncryptAlgo = algo
	transpEnc.EncParams.InitVector = hex.EncodeToString(initVect)

	return cipher.StreamWriter{
		S: cipherConstructor(aesEnc, initVect),
		W: NoopWtrCloser{dstWtr}, //Sheild our writer from being closed
	}, nil
}
