package mcp

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestReadRPCMessageContentLengthFrame(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
	r := bufio.NewReader(strings.NewReader(input))

	req, err := readRPCMessage(r)
	if err != nil {
		t.Fatalf("readRPCMessage failed: %v", err)
	}
	if req.Method != "initialize" {
		t.Fatalf("unexpected method: %s", req.Method)
	}
}

func TestReadRPCMessageLineDelimitedJSON(t *testing.T) {
	input := "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{}}\n"
	r := bufio.NewReader(strings.NewReader(input))

	req, err := readRPCMessage(r)
	if err != nil {
		t.Fatalf("readRPCMessage failed: %v", err)
	}
	if req.Method != "initialize" {
		t.Fatalf("unexpected method: %s", req.Method)
	}
}
