package lsp

import "testing"

func TestErrorResult(t *testing.T) {
	response := NewErrorResponse(1, RequestFailed, "We only support UTF-8 encoding :c")
	if response.Result != nil {
		t.Error("Result MUST be nil.")
	}
	if response.Error == nil {
		t.Errorf("Response.Error should be non-nil.")
	}
}
