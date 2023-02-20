package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Wrapper interface {
	Wait()
	Close()
}

func NewWs(addr string, handlers map[string]Handler) Wrapper {
	ws := &Ws{}

	ws.init(addr, handlers)

	return ws
}

type Ws struct {
	wg sync.WaitGroup
	s  *http.Server
}

type Handler func(route string, conn *websocket.Conn)

func (impl *Ws) init(addr string, handlers map[string]Handler) {
	r := mux.NewRouter()

	for s, handler := range handlers {
		r.HandleFunc(s, func(writer http.ResponseWriter, request *http.Request) {
			upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
				return true
			}}

			conn, err := upgrader.Upgrade(writer, request, nil)
			defer conn.Close()

			if err != nil {
				_, _ = writer.Write([]byte(err.Error()))

				return
			}

			handler(s, conn)
		})
	}

	s := &http.Server{
		Addr:        addr,
		Handler:     r,
		ReadTimeout: time.Second * 5,
	}

	impl.s = s

	impl.wg.Add(1)

	go func() {
		defer impl.wg.Done()

		if err := s.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
}

func (impl *Ws) Close() {
	_ = impl.s.Close()
}

func (impl *Ws) Wait() {
	impl.wg.Wait()
}
