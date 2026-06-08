package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type Document struct {
	Text    string
	Version int
}

type Server struct {
	mu          sync.RWMutex
	docs        map[string]*Document
	writer      io.Writer
	writeMu     sync.Mutex
	version     string
	rootURI     string
	initialized bool
	shutdown    bool
}

func Serve(r io.Reader, w io.Writer, version string) error {
	s := &Server{
		docs:    make(map[string]*Document),
		writer:  w,
		version: version,
	}
	reader := bufio.NewReader(r)
	for {
		raw, err := readMessage(reader)
		if err == io.EOF {
			if s.shutdown {
				os.Exit(0)
			}
			os.Exit(1)
		}
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		var msg RequestMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		s.dispatch(msg)
	}
}

func (s *Server) sendResponse(id json.RawMessage, result any, rpcErr *ResponseError) {
	resp := ResponseMessage{JSONRPC: "2.0", ID: id, Error: rpcErr}
	if rpcErr == nil {
		if result == nil {
			resp.Result = json.RawMessage("null")
		} else {
			resp.Result, _ = json.Marshal(result)
		}
	}
	body, _ := json.Marshal(resp)
	writeMessage(s.writer, &s.writeMu, body) //nolint:errcheck
}

func (s *Server) sendNotification(method string, params any) {
	notif := NotificationMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	body, _ := json.Marshal(notif)
	writeMessage(s.writer, &s.writeMu, body) //nolint:errcheck
}
