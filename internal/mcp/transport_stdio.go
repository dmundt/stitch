package mcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id,omitempty"`
	Result  any        `json:"result,omitempty"`
	Error   *rpcErrObj `json:"error,omitempty"`
}

type rpcErrObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func (a *app) serveSTDIO() error {
	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)

	for {
		req, err := readRPCMessage(r)
		if err != nil {
			return err
		}
		resp := a.handleRPC(req)
		if err := writeRPCMessage(w, resp); err != nil {
			return err
		}
	}
}

func readRPCMessage(r *bufio.Reader) (rpcRequest, error) {
	contentLen := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return rpcRequest{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if contentLen < 0 {
				continue
			}
			break
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "content-length:") {
			raw := strings.TrimSpace(line[len("content-length:"):])
			n, err := strconv.Atoi(raw)
			if err != nil {
				return rpcRequest{}, fmt.Errorf("invalid content-length: %w", err)
			}
			contentLen = n
		}
	}
	if contentLen <= 0 {
		return rpcRequest{}, errors.New("missing content-length")
	}
	body := make([]byte, contentLen)
	if _, err := io.ReadFull(r, body); err != nil {
		return rpcRequest{}, err
	}
	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return rpcRequest{}, err
	}
	return req, nil
}

func writeRPCMessage(w *bufio.Writer, resp rpcResponse) error {
	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := w.WriteString(header); err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	return w.Flush()
}

func (a *app) handleRPC(req rpcRequest) rpcResponse {
	id := req.ID
	if req.JSONRPC == "" {
		req.JSONRPC = "2.0"
	}
	switch req.Method {
	case "initialize":
		return rpcResponse{JSONRPC: "2.0", ID: id, Result: map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "stitch-mcp",
				"version": "0.1.0",
			},
		}}
	case "notifications/initialized":
		return rpcResponse{JSONRPC: "2.0", ID: id, Result: map[string]any{}}
	case "tools/list":
		return rpcResponse{JSONRPC: "2.0", ID: id, Result: map[string]any{"tools": mcpTools()}}
	case "tools/call":
		res, err := a.handleToolCall(req.Params)
		if err != nil {
			return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcErrObj{Code: -32000, Message: err.Error()}}
		}
		return rpcResponse{JSONRPC: "2.0", ID: id, Result: res}
	default:
		return rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcErrObj{Code: -32601, Message: "method not found"}}
	}
}

func (a *app) handleToolCall(raw json.RawMessage) (map[string]any, error) {
	var req toolCallRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	if req.Arguments == nil {
		req.Arguments = map[string]any{}
	}

	result, err := a.safeDispatch(req.Name, req.Arguments)
	if err != nil {
		return map[string]any{
			"isError": true,
			"content": []map[string]any{{"type": "text", "text": err.Error()}},
		}, nil
	}

	encoded, _ := json.MarshalIndent(result, "", "  ")
	return map[string]any{
		"isError":           false,
		"structuredContent": result,
		"content":           []map[string]any{{"type": "text", "text": string(encoded)}},
	}, nil
}
