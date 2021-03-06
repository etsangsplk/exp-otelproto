package ws_stream_async

import (
	"log"
	"net/http"

	"github.com/tigrannajaryan/exp-otelproto/encodings"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"

	"github.com/tigrannajaryan/exp-otelproto/core"
	"github.com/tigrannajaryan/exp-otelproto/encodings/otlp"
)

type Server struct {
}

var upgrader = websocket.Upgrader{} // use default options

func telemetryReceiver(w http.ResponseWriter, r *http.Request, onReceive func(batch core.ExportRequest, spanCount int)) {
	//log.Printf("Incoming WS connection.")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("upgrade:", err)
		return
	}
	defer c.Close()
	lastId := uint64(0)
	for {
		mt, bytes, err := c.ReadMessage()
		if err != nil {
			log.Fatal("read:", err)
			break
		}

		request := encodings.Decode(bytes)

		Id := request.GetExport().Id
		if Id == 0 {
			log.Fatal("Received 0 Id")
		}
		if Id != lastId+1 {
			log.Fatalf("Received out of order request ID=%d", Id)
		}
		lastId = Id

		onReceive(request, len(request.GetExport().ResourceSpans[0].Spans))

		response := &otlp.Response{
			Body: &otlp.Response_Export{
				Export: &otlp.ExportResponse{Id: Id},
			},
		}
		responseBytes, err := proto.Marshal(response)
		if err != nil {
			log.Fatal("cannot encode:", err)
			break
		}

		err = c.WriteMessage(mt, responseBytes)
		if err != nil {
			log.Fatal("write:", err)
			break
		}
	}
}

func (srv *Server) Listen(endpoint string, onReceive func(batch core.ExportRequest, spanCount int)) error {
	http.HandleFunc(
		"/telemetry",
		func(w http.ResponseWriter, r *http.Request) { telemetryReceiver(w, r, onReceive) },
	)
	log.Fatal(http.ListenAndServe(endpoint, nil))
	return nil
}

func (srv *Server) Stop() {
}
