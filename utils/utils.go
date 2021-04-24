package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/constants"
)

func Uuid() (id string) {
	idByte := make([]byte, 16)

	_, err := rand.Read(idByte)
	if err != nil {
		err = &IoError{
			errors.Wrap(err, "utils: Failed to get random data"),
		}
		panic(err)
	}

	id = hex.EncodeToString(idByte[:])

	return
}

func GetRootDir() (pth string) {
	pth, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	return
}

func GetLogPath() (pth string) {
	if constants.Development {
		pth = filepath.Join(GetRootDir(), "log")

		err := os.MkdirAll(pth, 0755)
		if err != nil {
			err = &IoError{
				errors.Wrap(err, "utils: Failed to create dev directory"),
			}
			panic(err)
		}

		pth = filepath.Join(pth, "pritunl_endpoint.log")

		return
	}

	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join("C:\\", "ProgramData", "Pritunl")

		err := os.MkdirAll(pth, 0755)
		if err != nil {
			err = &IoError{
				errors.Wrap(err, "utils: Failed to create data directory"),
			}
			panic(err)
		}

		pth = filepath.Join(pth, "pritunl_endpoint.log")
		break
	case "linux", "darwin":
		pth = filepath.Join(string(filepath.Separator),
			"var", "log", "pritunl_endpoint.log")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}

func GetLogPath2() (pth string) {
	if constants.Development {
		pth = filepath.Join(GetRootDir(), "log")

		err := os.MkdirAll(pth, 0755)
		if err != nil {
			err = &IoError{
				errors.Wrap(err, "utils: Failed to create dev directory"),
			}
			panic(err)
		}

		pth = filepath.Join(pth, "pritunl_endpoint.log.1")

		return
	}

	switch runtime.GOOS {
	case "windows":
		pth = filepath.Join("C:\\", "ProgramData", "Pritunl")

		err := os.MkdirAll(pth, 0755)
		if err != nil {
			err = &IoError{
				errors.Wrap(err, "utils: Failed to create data directory"),
			}
			panic(err)
		}

		pth = filepath.Join(pth, "pritunl_endpoint.log.1")
		break
	case "linux", "darwin":
		pth = filepath.Join(string(filepath.Separator),
			"var", "log", "pritunl_endpoint.log.1")
		break
	default:
		panic("profile: Not implemented")
	}

	return
}
