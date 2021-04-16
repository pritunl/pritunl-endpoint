package stream

import (
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/sirupsen/logrus"
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

func (s *Stream) Run() {
	for {
		doc := <-s.primary

		doc.Print()
	}
}

func New() (strm *Stream) {
	return &Stream{
		primary:   make(chan Doc, 1024),
		secondary: make(chan Doc, 1024),
	}
}
