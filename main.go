package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/caleb-fringer/imp_lsp/internal/analysis"
	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/rpc"
)

func main() {
	logger, err := getLogger("/home/caleb/src/imp_lsp/log.txt")
	if err != nil {
		log.Fatalf("Couldn't open log file: %v", err)
	}
	logger.Println("Hi :)")

	serverState, err := analysis.NewState(logger)
	if err != nil {
		log.Fatalf("Couldnt initialize server state: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.Split)

	for scanner.Scan() {
		msg := scanner.Bytes()
		method, contents, err := rpc.DecodeMessage(msg)
		if err != nil {
			logger.Printf("Error decoding message: %v\n", err)
			continue
		}
		handleMessage(logger, method, contents, serverState)
	}
}

func getLogger(filename string) (*log.Logger, error) {
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	return log.New(logfile, "[imp_lsp] ", log.Ldate|log.Ltime|log.Lshortfile), nil
}

func handleMessage(logger *log.Logger, method string, contents []byte, state *analysis.ServerState) {
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
		message := lsp.NewInitializeResponse(request.ID, lsp.Full)
		response := rpc.EncodeMessage(message)
		os.Stdout.WriteString(response)
		logger.Println("Replied! :)")

	case "textDocument/didOpen":
		var notification lsp.DidOpenTextDocumentNotification
		err := json.Unmarshal(contents, &notification)
		if err != nil {
			logger.Printf("Error unmarshalling DidOpenTextDocumentNotification: %v\n", err)
			return
		}
		logger.Printf("Opened: %s\n", notification.Params.TextDocument.URI)
		// Parse the document
		_, err = state.OpenDocument(&notification.Params.TextDocument)
		if err != nil {
			logger.Printf("Error opening document: %v\n", err)
		}

		// Create a DUMMY diagnostic for testing
		diagnostic := lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      0,
					Character: 1,
				},
			},
			Severity: lsp.Information,
			Source:   "imp_lsp",
			Message:  "Good Job! :)",
		}
		diagnosticsNotification := lsp.NewPublishDiagnosticsNotification([]lsp.Diagnostic{diagnostic}, notification.Params.TextDocument.URI, 1)
		response := rpc.EncodeMessage(diagnosticsNotification)
		logger.Printf("Writing the following diagnostics:\n\t%v\n", notification)
		os.Stdout.WriteString(response)

	case "textDocument/didChange":
		// Update state of document
		var notification lsp.DidChangeTextDocumentNotification
		err := json.Unmarshal(contents, &notification)
		if err != nil {
			logger.Printf("Error unmarshalling DidChangeTextDocumentNotification: %v\n", err)
			return
		}
		logger.Printf("Edited: %s\n", notification.Params.TextDocument.URI)
		_, err = state.EditDocument(&notification)
		if err != nil {
			logger.Printf("Error editing document: %v\n", err)
		}
		// TODO: Run new diagnostics...
	}
}
