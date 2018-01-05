package transpenc

import (
	"bytes"
	"testing"
)

func TestTranportEncoding(t *testing.T) {
	origEnc := &TranportEncoding{
		CompressAlgo: "TestCompressAlgo",
		EncParams: EncryptionParams{
			KeyName:     "TestKeyName",
			EncryptAlgo: "TestEncryptAlgo",
			InitVector:  "TestInitVector",
		},
	}
	encBuf := new(bytes.Buffer)
	origEnc.Write(encBuf)
	resultEnc := new(TranportEncoding)

	if err := ReadTranportEncoding(bytes.NewReader(encBuf.Bytes()), resultEnc); err != nil {
		t.Fatalf("Unexepected error while testing TranportEncoding deserialization: %s", err)
	}
	if *origEnc != *resultEnc {
		t.Fatalf("Deserialized TranportEncoding did not match original TranportEncoding that was serialized!")
	}
}
