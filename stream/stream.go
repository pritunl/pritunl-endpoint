package stream

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/websocket"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-endpoint/constants"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/sirupsen/logrus"
)

var (
	Dialer = &websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}
)

const (
	endpointWriteTimeout = 10 * time.Second
	endpointPingInterval = 30 * time.Second
	endpointPingWait     = 40 * time.Second
)

type Stream struct {
	primary   chan Doc
	secondary chan Doc
}

type Doc interface {
	GetId() primitive.ObjectID
	SetId(primitive.ObjectID)
	GetTimestamp() time.Time
	SetTimestamp(time.Time)
	GetType() string
	Print()
}

func (s *Stream) Append(doc Doc) {
	doc.SetId(primitive.NewObjectID())
	doc.SetTimestamp(time.Now())

	if len(s.primary) > 1000 {
		logrus.WithFields(logrus.Fields{
			"length": len(s.primary),
		}).Error("stream: Stream buffer full")
		return
	}

	s.primary <- doc
}

func WriteDoc(conn *websocket.Conn, doc Doc) (err error) {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "mhandlers: Failed to get writer"),
		}
		return
	}
	defer w.Close()

	_, err = w.Write([]byte("system:"))
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "mhandlers: Failed to write prefix"),
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
	streamUrl, err := url.Parse(constants.TodoStreamUrl)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "stream: Failed to parse stream URL"),
		}
		return
	}

	conn, _, err := Dialer.Dial(streamUrl.String(), nil)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "stream: Failed to dial stream"),
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

	ticker := time.NewTicker(endpointPingWait)

	go func() {
		defer func() {
			recover()
		}()
		for {
			_, _, err := conn.NextReader()
			if err != nil {
				conn.Close()
				break
			}
		}
	}()

	for {
		select {
		case doc, ok := <-s.primary:
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
				err = &errortypes.RequestError{
					errors.Wrap(err,
						"mhandlers: Failed to set write deadline"),
				}
				return
			}

			err = WriteDoc(conn, doc)
			if err != nil {
				err = &errortypes.RequestError{
					errors.Wrap(err,
						"mhandlers: Failed to set write json"),
				}
				return
			}
		case <-ticker.C:
			err = conn.WriteControl(websocket.PingMessage, []byte{},
				time.Now().Add(endpointWriteTimeout))
			if err != nil {
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

		time.Sleep(1 * time.Second)
	}
}

func New() (strm *Stream) {
	return &Stream{
		primary:   make(chan Doc, 1024),
		secondary: make(chan Doc, 1024),
	}
}
