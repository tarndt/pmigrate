package transpenc

import (
	"encoding/binary"
	"encoding/json"
	"io"

	"lib/errs"
)

type EncryptionParams struct {
	KeyName     string
	EncryptAlgo string
	InitVector  string
}

type TranportEncoding struct {
	CompressAlgo string
	EncParams    EncryptionParams
}

func (this TranportEncoding) Write(wtr io.Writer) error {
	rawBytes, err := json.Marshal(&this)
	if err != nil {
		return errs.Append(err, "Could not marshal transport encoding parameters")
	}
	if err = binary.Write(wtr, binary.LittleEndian, uint32(len(rawBytes))); err != nil {
		return errs.Append(err, "Could not write transport encoding length")
	}
	if _, err = wtr.Write(rawBytes); err != nil {
		return errs.Append(err, "Could not write transport encoding parameters")
	}
	return nil
}

func ReadTranportEncoding(rdr io.Reader, transpEnc *TranportEncoding) error {
	var length uint32
	err := binary.Read(rdr, binary.LittleEndian, &length)
	if err != nil {
		return errs.Append(err, "Could not read transport encoding length")
	}
	rawBytes := make([]byte, length)
	if _, err = io.ReadFull(rdr, rawBytes); err != nil {
		return errs.Append(err, "Could not read transport encoding parameters")
	}
	if err = json.Unmarshal(rawBytes, transpEnc); err != nil {
		return errs.Append(err, "Could not unmarshal transport encoding parameters")
	}
	return nil
}
