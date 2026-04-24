package lsp

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// TestDecoder_SingleMessages uses table-driven testing to verify
// how the decoder handles individual valid and malformed messages.
func TestDecoder_SingleMessages(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMsg *Message
		wantErr bool
	}{
		{
			name:  "Valid message with Content-Type",
			input: "Content-Length: 17\r\nContent-Type: application/json\r\n\r\n{\"valid\": \"json\"}",
			wantMsg: &Message{
				ContentLength: 17,
				ContentType:   "application/json",
				Body:          map[string]any{"valid": "json"},
			},
			wantErr: false,
		},
		{
			name:  "Valid message without Content-Type",
			input: "Content-Length: 17\r\n\r\n{\"valid\": \"json\"}",
			wantMsg: &Message{
				ContentLength: 17,
				ContentType:   "",
				Body:          map[string]any{"valid": "json"},
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON body",
			input:   "Content-Length: 10\r\n\r\n{bad-json}",
			wantMsg: nil,
			wantErr: true,
		},
		{
			name:    "Malformed header name",
			input:   "Content-Length: 17\r\nBad-Header: value\r\n\r\n{\"valid\": \"json\"}",
			wantMsg: nil,
			wantErr: true,
		},
		{
			name:    "Missing \r in delimiter (strict mode failure)",
			input:   "Content-Length: 17\n\n{\"valid\": \"json\"}",
			wantMsg: nil,
			wantErr: true, // Will hit EOF trying to find a strict \r\n boundary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDecoder(strings.NewReader(tt.input))
			msg, err := decoder.Decode()

			if (err != nil) != tt.wantErr {
				t.Fatalf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			// reflect.DeepEqual is perfect for comparing structs that contain maps
			if !reflect.DeepEqual(msg, tt.wantMsg) {
				t.Errorf("Decode() = %+v, want %+v", msg, tt.wantMsg)
			}
		})
	}
}

// TestDecoder_ResyncSequence verifies that if a message fails mid-stream,
// the decoder successfully finds the next valid message boundary.
func TestDecoder_ResyncSequence(t *testing.T) {
	// A continuous stream containing:
	// 1. Valid message
	// 2. Malformed message (Bad header)
	// 3. Valid message

	input := "Content-Length: 12\r\n\r\n{\"msg\": \"1\"}" +
		"Content-Length: 15\r\nBad-Header: foo\r\n\r\n{\"bad\":\"data\"}" +
		"Content-Length: 12\r\n\r\n{\"msg\": \"3\"}"

	decoder := NewDecoder(strings.NewReader(input))

	// 1. First message should succeed
	msg1, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Expected message 1 to succeed, got error: %v", err)
	}
	if msg1.Body["msg"] != "1" {
		t.Errorf("Expected message 1 body 'msg' to be '1', got %v", msg1.Body["msg"])
	}

	// 2. Second message should fail (because of Bad-Header)
	_, err = decoder.Decode()
	if err == nil {
		t.Fatalf("Expected message 2 to fail due to bad header, but it succeeded")
	}

	// 3. Third message should succeed (proving the stream successfully resynced)
	msg3, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Expected message 3 to succeed after resync, got error: %v", err)
	}
	if msg3.Body["msg"] != "3" {
		t.Errorf("Expected message 3 body 'msg' to be '3', got %v", msg3.Body["msg"])
	}

	// 4. Checking for proper EOF handling
	_, err = decoder.Decode()
	if !errors.Is(err, io.EOF) {
		t.Errorf("Expected io.EOF at end of stream, got: %v", err)
	}
}
