package nonce

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
)

var nonces = set.NewSet()

func Validate(nce string) (err error) {
	if nonces.Contains(nce) {
		err = &errortypes.AuthenticationError{
			errors.New("nonce: Duplicate authentication nonce"),
		}
	}

	nonces.Add(nce)

	return
}
