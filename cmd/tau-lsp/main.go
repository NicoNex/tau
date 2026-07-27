// Command tau-lsp is the language server for Tau. It speaks LSP over stdio:
// Content-Length framed JSON-RPC 2.0 on stdin and stdout, which is all any
// editor asks for. The analysis behind it is the compiler's own lexer,
// parser and formatter, so what the server reports is what the compiler
// would report and nothing has to be kept in sync by hand.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// JSON-RPC error codes used by this server.
const (
	errParse          = -32700
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInternal       = -32603
	errServerNotInit  = -32002
)

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result"`
	Error   *responseError  `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// conn is the framed JSON-RPC transport.
type conn struct {
	in  *bufio.Reader
	out *bufio.Writer
}

// read returns the body of the next message. A frame without a usable
// Content-Length is not something we can resynchronise from, so it ends the
// stream rather than leaving the reader at an unknown offset.
func (c *conn) read() ([]byte, error) {
	length := -1

	for {
		line, err := c.in.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("bad Content-Length %q", value)
			}
			length = n
		}
	}

	if length < 0 {
		return nil, errors.New("message without Content-Length")
	}
	if length == 0 {
		return []byte{}, nil
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(c.in, body); err != nil {
		return nil, err
	}
	return body, nil
}

func (c *conn) write(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("tau-lsp: marshal: %v", err)
		return
	}
	fmt.Fprintf(c.out, "Content-Length: %d\r\n\r\n", len(data))
	c.out.Write(data)
	c.out.Flush()
}

func (c *conn) reply(id json.RawMessage, result any) {
	c.write(response{JSONRPC: "2.0", ID: id, Result: result})
}

func (c *conn) replyErr(id json.RawMessage, code int, msg string) {
	c.write(response{JSONRPC: "2.0", ID: id, Error: &responseError{Code: code, Message: msg}})
}

func (c *conn) notify(method string, params any) {
	c.write(notification{JSONRPC: "2.0", Method: method, Params: params})
}

// isRequest tells a request, which must be answered, from a notification,
// which must not. A null id is a notification too.
func isRequest(id json.RawMessage) bool {
	s := strings.TrimSpace(string(id))
	return s != "" && s != "null"
}

func main() {
	logfile := flag.String("log", "", "write the server log to this file")
	flag.Parse()

	log.SetFlags(log.Ltime)
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("tau-lsp: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		// Never stdout: that channel belongs to the protocol.
		log.SetOutput(io.Discard)
	}

	s := newServer(&conn{
		in:  bufio.NewReader(os.Stdin),
		out: bufio.NewWriter(os.Stdout),
	})

	if err := s.run(); err != nil && err != io.EOF {
		log.Printf("tau-lsp: %v", err)
		os.Exit(1)
	}
}
