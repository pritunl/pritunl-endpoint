package stream

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/websocket"
	"github.com/pritunl/pritunl-endpoint/config"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/utils"
	"github.com/sirupsen/logrus"
)

var (
	Dialer = &websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}
)

const (
	primaryBufferSize    = 1100
	secondaryBufferSize  = 50100
	writingBufferSize    = 1100
	writingSwap          = 3 * time.Minute
	endpointWriteTimeout = 10 * time.Second
	endpointPingInterval = 30 * time.Second
	endpointPingWait     = 40 * time.Second
)

type Stream struct {
	primary          chan Doc
	secondary        chan Doc
	writingTimestamp time.Time
	writingPresent   chan Doc
	writingPast      chan Doc
}

type Doc interface {
	GetTimestamp() time.Time
	SetTimestamp(time.Time)
	GetType() string
	Print()
}

func (s *Stream) Append(doc Doc) {
	doc.SetTimestamp(time.Now().UTC())

	if len(s.primary) > primaryBufferSize-100 {
		s.AppendSecondary(doc)
		return
	}

	s.primary <- doc
}

func (s *Stream) AppendSecondary(doc Doc) {
	if len(s.secondary) > secondaryBufferSize-100 {
		logrus.WithFields(logrus.Fields{
			"length": len(s.secondary),
		}).Error("stream: Buffer full, dropping doc")
		return
	}

	s.secondary <- doc
}

func (s *Stream) AppendWriting(doc Doc) {
	if len(s.writingPresent) > writingBufferSize-100 ||
		time.Since(s.writingTimestamp) > writingSwap {

		s.writingTimestamp = time.Now()
		s.writingPast = s.writingPresent
		s.writingPresent = make(chan Doc, writingBufferSize)
	}

	s.writingPresent <- doc
}

func (s *Stream) RecoverBuffer() {
	close(s.writingPresent)
	close(s.writingPast)

	for doc := range s.writingPresent {
		if len(s.secondary) > secondaryBufferSize-100 {
			logrus.WithFields(logrus.Fields{
				"length": len(s.secondary),
			}).Error("stream: Buffer full on recover present, dropping doc")
			break
		}

		s.secondary <- doc
	}
	for doc := range s.writingPast {
		if len(s.secondary) > secondaryBufferSize-100 {
			logrus.WithFields(logrus.Fields{
				"length": len(s.secondary),
			}).Error("stream: Buffer full on recover past, dropping doc")
			break
		}

		s.secondary <- doc
	}

	s.writingTimestamp = time.Now()
	s.writingPast = make(chan Doc, writingBufferSize)
	s.writingPresent = make(chan Doc, writingBufferSize)
}

func WriteDoc(conn *websocket.Conn, doc Doc) (err error) {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "stream: Failed to get writer"),
		}
		return
	}
	defer w.Close()

	_, err = w.Write([]byte(doc.GetType() + ":"))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "stream: Failed to write prefix"),
		}
		return
	}

	err = json.NewEncoder(w).Encode(doc)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "mhandlers: Failed to write json"),
		}
		return
	}

	return
}

func (s *Stream) Conn() (err error) {
	streamUrl := &url.URL{
		Scheme: "wss",
		Host:   config.Config.RemoteHost,
		Path:   fmt.Sprintf("/endpoint/%s/comm", config.Config.Id),
	}

	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	nonce, err := utils.RandStr(64)
	if err != nil {
		return
	}

	authString := strings.Join([]string{
		timestampStr,
		nonce,
		"communicate",
	}, "&")

	hashFunc := hmac.New(sha512.New, []byte(config.Config.Secret))
	hashFunc.Write([]byte(authString))
	rawSignature := hashFunc.Sum(nil)
	signature := base64.URLEncoding.EncodeToString(rawSignature)

	header := http.Header{}
	header.Add("Pritunl-Endpoint-Timestamp", timestampStr)
	header.Add("Pritunl-Endpoint-Nonce", nonce)
	header.Add("Pritunl-Endpoint-Signature", signature)

	conn, res, err := Dialer.Dial(streamUrl.String(), header)
	if err != nil {
		if res != nil {
			errData := &errortypes.ErrorData{}
			e := json.NewDecoder(res.Body).Decode(errData)
			if e == nil {
				logrus.WithFields(logrus.Fields{
					"error_code": errData.Error,
					"error_msg":  errData.Message,
				}).Error("endpoint: Communicate error")
			}
			err = &errortypes.ConnectionError{
				errors.Wrapf(
					err,
					"stream: Failed to dial stream, status '%d'",
					res.StatusCode,
				),
			}
		} else {
			err = &errortypes.ConnectionError{
				errors.Wrap(err, "stream: Failed to dial stream"),
			}
		}

		return
	}
	defer conn.Close()

	err = conn.SetReadDeadline(time.Now().Add(endpointPingWait))
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "stream: Failed to set read deadline"),
		}
		return
	}

	conn.SetPongHandler(func(x string) (err error) {
		err = conn.SetReadDeadline(time.Now().Add(endpointPingWait))
		if err != nil {
			err = &errortypes.RequestError{
				errors.Wrap(err, "stream: Failed to set read deadline"),
			}
			return
		}

		return
	})

	ticker := time.NewTicker(endpointPingInterval)

	go func() {
		defer func() {
			recover()
		}()
		for {
			_, _, err := conn.NextReader()
			if err != nil {
				conn.Close()
				return
			}
		}
	}()

	for {
		select {
		case doc, ok := <-s.primary:
			s.AppendWriting(doc)

			if !ok {
				err = conn.WriteControl(websocket.CloseMessage, []byte{},
					time.Now().Add(endpointWriteTimeout))
				if err != nil {
					err = &errortypes.RequestError{
						errors.Wrap(err,
							"mhandlers: Failed to set write control"),
					}
					return
				}
				return
			}

			err = conn.SetWriteDeadline(time.Now().Add(endpointWriteTimeout))
			if err != nil {
				s.AppendSecondary(doc)
				err = &errortypes.RequestError{
					errors.Wrap(err,
						"mhandlers: Failed to set write deadline"),
				}
				return
			}

			err = WriteDoc(conn, doc)
			if err != nil {
				s.AppendSecondary(doc)
				err = &errortypes.RequestError{
					errors.Wrap(err,
						"mhandlers: Failed to set write json"),
				}
				return
			}
		case doc, ok := <-s.secondary:
			s.AppendWriting(doc)

			if !ok {
				err = conn.WriteControl(websocket.CloseMessage, []byte{},
					time.Now().Add(endpointWriteTimeout))
				if err != nil {
					err = &errortypes.RequestError{
						errors.Wrap(err,
							"mhandlers: Failed to set write control"),
					}
					return
				}
				return
			}

			err = conn.SetWriteDeadline(time.Now().Add(endpointWriteTimeout))
			if err != nil {
				s.AppendSecondary(doc)
				err = &errortypes.RequestError{
					errors.Wrap(err,
						"mhandlers: Failed to set write secondary deadline"),
				}
				return
			}

			err = WriteDoc(conn, doc)
			if err != nil {
				s.AppendSecondary(doc)
				err = &errortypes.RequestError{
					errors.Wrap(err,
						"mhandlers: Failed to set write secondary json"),
				}
				return
			}
		case <-ticker.C:
			err = conn.WriteControl(websocket.PingMessage, []byte{},
				time.Now().Add(endpointWriteTimeout))
			if err != nil {
				err = &errortypes.RequestError{
					errors.Wrap(err,
						"mhandlers: Failed to set write control"),
				}
				return
			}
		}
	}
}

func (s *Stream) Run() {
	for {
		err := s.Conn()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("stream: Stream conn error")
		}

		s.RecoverBuffer()

		time.Sleep(1 * time.Second)
	}
}

func New() (strm *Stream) {
	return &Stream{
		primary:          make(chan Doc, primaryBufferSize),
		secondary:        make(chan Doc, secondaryBufferSize),
		writingTimestamp: time.Now(),
		writingPresent:   make(chan Doc, writingBufferSize),
		writingPast:      make(chan Doc, writingBufferSize),
	}
}
