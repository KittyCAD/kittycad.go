package kittycad_test

import (
	"encoding/json"
	"testing"

	kittycad "github.com/kittycad/kittycad.go"
)

func TestLegacyCopilotUserRetainsPayload(t *testing.T) {
	// A downstream initializer using the previous SDK's public name must still
	// describe a user message after regeneration, not a system command.
	message := kittycad.MlCopilotClientMessageCurrentFiles{
		Type: "user", Content: "make a cube", Mode: kittycad.MlCopilotModeFast,
		CurrentFiles: map[string][]int{"main.kcl": {65}},
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	var decoded kittycad.MlCopilotMessageUser
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Content != "make a cube" || decoded.Type != "user" || decoded.Mode != "fast" || len(decoded.CurrentFiles["main.kcl"]) != 1 {
		t.Fatalf("lost user payload: %s", encoded)
	}
	var legacy kittycad.MlCopilotClientMessageCurrentFiles
	if err := json.Unmarshal([]byte(`{"type":"user","content":"existing message","current_files":{"main.kcl":[65]}}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Content != "existing message" || legacy.CurrentFiles["main.kcl"][0] != 65 {
		t.Fatal("legacy decoding lost content")
	}
}

func TestLegacyCopilotAliasesKeepTheirMeaning(t *testing.T) {
	var _ kittycad.MlCopilotMessageSystem = kittycad.MlCopilotClientMessageProjectName{Type: "system"}
	var _ kittycad.MlCopilotMessageAttachmentResponse = kittycad.MlCopilotClientMessageProjectContext{Type: "attachment_response"}
	var _ kittycad.MlCopilotMessageProjectContext = kittycad.MlCopilotClientMessageMlCopilotClientMessageHeaders{Type: "project_context"}
	var _ kittycad.MlCopilotMessageHeaders = kittycad.MlCopilotClientMessageHeaders{Type: "headers"}
	var _ kittycad.MlCopilotMessageListModes = kittycad.MlCopilotClientMessageListModes{Type: "list_modes"}
	var _ kittycad.MlCopilotMessagePing = kittycad.MlCopilotClientMessagePing{Type: "ping"}
}
