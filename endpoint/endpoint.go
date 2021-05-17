package endpoint

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/config"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/nonce"
	"github.com/pritunl/pritunl-endpoint/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
)

var (
	clientTransport = &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		},
	}
	client = &http.Client{
		Transport: clientTransport,
		Timeout:   10 * time.Second,
	}
)

type RegisterData struct {
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

func GenerateKey() (pubKey64, privKey64 string, err error) {
	pubKey, privKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "endpoint: Failed to generate nacl key"),
		}
		return
	}

	pubKey64 = base64.StdEncoding.EncodeToString(pubKey[:])
	privKey64 = base64.StdEncoding.EncodeToString(privKey[:])

	return
}

func Register() (err error) {
	pubKey, privKey, err := GenerateKey()
	if err != nil {
		return
	}

	regNonce, err := utils.RandStr(64)
	if err != nil {
		return
	}

	regData := &RegisterData{
		Timestamp: time.Now().Unix(),
		Nonce:     regNonce,
		PublicKey: pubKey,
	}

	authString := strings.Join([]string{
		strconv.FormatInt(regData.Timestamp, 10),
		regData.Nonce,
		regData.PublicKey,
	}, "&")

	hashFunc := hmac.New(sha512.New, []byte(config.Config.Secret))
	hashFunc.Write([]byte(authString))
	rawSignature := hashFunc.Sum(nil)
	regData.Signature = base64.StdEncoding.EncodeToString(rawSignature)

	reqData, err := json.Marshal(regData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "endpoint: Failed to marshal data"),
		}
		return
	}

	u := &url.URL{
		Scheme: "https",
		Host:   config.Config.RemoteHost,
		Path:   fmt.Sprintf("/endpoint/%s/register", config.Config.Id),
	}

	req, err := http.NewRequest(
		"PUT",
		u.String(),
		bytes.NewBuffer(reqData),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "endpoint: Request put error"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-endpoint")
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "endpoint: Request put error"),
		}
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			logrus.WithFields(logrus.Fields{
				"error_code": "endpoint_not_found",
				"error_msg":  "Endpoint does not exist",
			}).Error("endpoint: Register error")
		} else if res.StatusCode >= 400 && res.StatusCode < 500 {
			errData := &errortypes.ErrorData{}
			e := json.NewDecoder(res.Body).Decode(errData)
			if e == nil {
				logrus.WithFields(logrus.Fields{
					"error_code": errData.Error,
					"error_msg":  errData.Message,
				}).Error("endpoint: Register error")
			}
		}

		err = &errortypes.RequestError{
			errors.Wrapf(err, "endpoint: Bad status %n code from server",
				res.StatusCode),
		}
		return
	}

	resData := &RegisterData{}
	err = json.NewDecoder(res.Body).Decode(resData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "endpoint: Failed to parse response body"),
		}
		return
	}

	if len(resData.Nonce) < 16 || len(resData.Nonce) > 128 {
		err = &errortypes.AuthenticationError{
			errors.New("endpoint: Invalid authentication nonce"),
		}
		return
	}

	if len(resData.PublicKey) < 16 || len(resData.PublicKey) > 512 {
		err = &errortypes.AuthenticationError{
			errors.New("endpoint: Invalid public key"),
		}
		return
	}

	timestamp := time.Unix(resData.Timestamp, 0)
	if utils.SinceAbs(timestamp) > 300*time.Second {
		err = &errortypes.AuthenticationError{
			errors.New("endpoint: Authentication timestamp outside window"),
		}
		return
	}

	authString = strings.Join([]string{
		strconv.FormatInt(resData.Timestamp, 10),
		resData.Nonce,
		resData.PublicKey,
	}, "&")

	err = nonce.Validate(resData.Nonce)
	if err != nil {
		return
	}

	hashFunc = hmac.New(sha512.New, []byte(config.Config.Secret))
	hashFunc.Write([]byte(authString))
	rawSignature = hashFunc.Sum(nil)
	testSig := base64.StdEncoding.EncodeToString(rawSignature)

	if subtle.ConstantTimeCompare(
		[]byte(testSig), []byte(resData.Signature)) != 1 {

		err = &errortypes.ParseError{
			errors.Wrap(err, "endpoint: Response signature invalid"),
		}
		return
	}

	config.Config.PublicKey = pubKey
	config.Config.PrivateKey = privKey

	err = config.Save()
	if err != nil {
		return
	}

	return
}

func Init() (err error) {
	if config.Config.Id == "" {
		err = &errortypes.ParseError{
			errors.New("endpoint: Config missing ID"),
		}
		return
	}
	if config.Config.RemoteHost == "" {
		err = &errortypes.ParseError{
			errors.New("endpoint: Config missing remote host"),
		}
		return
	}
	if config.Config.Secret == "" {
		err = &errortypes.ParseError{
			errors.New("endpoint: Config missing secret"),
		}
		return
	}

	if config.Config.PublicKey != "" && config.Config.PrivateKey != "" {
		return
	}

	err = Register()
	if err != nil {
		return
	}

	return
}
