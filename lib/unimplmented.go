package lib

import (
	"errors"
)

var ErrFeatureUnimplemented = errors.New("Error: Feature is not yet implemented!")

func PanicUnimplemented() {
	panic(ErrFeatureUnimplemented.Error())
}
