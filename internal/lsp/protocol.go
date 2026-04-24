package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Message struct {
	ContentLength uint
	ContentType   string
	Body          map[string]any
}

// Decoder wraps an io.Reader to decode a stream of Messages.
type Decoder struct {
	br *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		br: bufio.NewReader(r),
	}
}

// Decode reads the next Message from the stream.
func (d *Decoder) Decode() (*Message, error) {
	msg := &Message{}

	// Phase 1: Resync and find Content-Length
	for {
		line, err := d.br.ReadString('\n')
		// EOF found or other error
		if err != nil {
			return nil, err
		}

		// A malformed message's unread body might precede the start of the next
		// valid message on the exact same line. Look for the signature anywhere.
		const prefix = "Content-Length: "
		idx := strings.Index(line, prefix)

		if idx != -1 {
			// Extract the substring immediately after "Content-Length: "
			start := idx + len(prefix)

			// TrimSpace automatically cleans up the \r\n, \n, or any accidental spaces
			numStr := strings.TrimSpace(line[start:])

			length, err := strconv.Atoi(numStr)
			// Header not found, continue parsing from the next line
			if err != nil {
				continue
			}

			msg.ContentLength = uint(length)
			break
		}
	}

	// Phase 2: Parse optional headers and find the empty \r\n line
	for {
		line, err := d.br.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading headers: %w", err)
		}

		// End of headers
		if line == "\r\n" {
			break
		}

		// Optional Content-Type header
		if strings.HasPrefix(line, "Content-Type: ") {
			msg.ContentType = strings.TrimSpace(line[len("Content-Type: "):])
		} else {
			// Reject the message if we encounter unexpected or malformed headers
			return nil, fmt.Errorf("unexpected or malformed header line: %q", line)
		}
	}

	// Phase 3: Read exact body bytes
	if msg.ContentLength > 0 {
		body := make([]byte, msg.ContentLength)
		_, err := io.ReadFull(d.br, body)
		if err != nil {
			return nil, fmt.Errorf("failed to read full body: %w", err)
		}

		// Phase 4: Parse JSON
		if err := json.Unmarshal(body, &msg.Body); err != nil {
			return nil, fmt.Errorf("invalid JSON body: %w", err)
		}
	}

	return msg, nil
}
