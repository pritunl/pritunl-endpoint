package stream

import (
	"fmt"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/sirupsen/logrus"
)

type Stream struct {
	primary   chan *Doc
	secondary chan *Doc
}

type Doc struct {
	Id        primitive.ObjectID     `json:"i"`
	Timestamp time.Time              `json:"t"`
	Type      string                 `json:"x"`
	Fields    map[string]interface{} `json:"f"`
}

func (s *Stream) Append(typ string, fields map[string]interface{}) {
	doc := &Doc{
		Id:        primitive.NewObjectID(),
		Timestamp: time.Now(),
		Type:      typ,
		Fields:    fields,
	}

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

		println("***************************************************")
		println(doc.Id.Hex())
		fmt.Println(doc.Timestamp)
		println(doc.Type)
		fmt.Println(doc.Fields)
		println("***************************************************")
	}
}

func New() (strm *Stream) {
	return &Stream{
		primary:   make(chan *Doc, 1024),
		secondary: make(chan *Doc, 1024),
	}
}
