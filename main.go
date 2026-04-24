package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/rpc"
)

func main() {
	logger, err := getLogger("/home/caleb/src/imp_lsp/log.txt")
	if err != nil {
		log.Fatalf("Couldn't open log file: %v", err)
	}
	logger.Println("Hi :)")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.Split)

	for scanner.Scan() {
		msg := scanner.Bytes()
		method, contents, err := rpc.DecodeMessage(msg)
		if err != nil {
			logger.Printf("Error decoding message: %v\n", err)
			continue
		}
		handleMessage(logger, method, contents)
	}
}

func getLogger(filename string) (*log.Logger, error) {
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	return log.New(logfile, "[imp_lsp] ", log.Ldate|log.Ltime|log.Lshortfile), nil
}

func handleMessage(logger *log.Logger, method string, contents []byte) {
	logger.Printf("Received %s message: %s\n", method, contents)
	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		err := json.Unmarshal(contents, &request)
		if err != nil {
			logger.Printf("Error unmarshalling InitializeRequest: %v\n", err)
			return
		}
		log.Printf("Connected to client: %s %s\n", request.Params.ClientInfo.Name, request.Params.ClientInfo.Version)
		message := lsp.NewInitializeResponse(request.ID)
		response := rpc.EncodeMessage(message)
		os.Stdout.WriteString(response)
		logger.Println("Replied! :)")
	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification
		err := json.Unmarshal(contents, &request)
		if err != nil {
			logger.Printf("Error unmarshalling DidOpenTextDocumentNotification: %v\n", err)
			return
		}
		logger.Printf("Opened: %s\n", request.Params.TextDocument.URI)
	}
}
