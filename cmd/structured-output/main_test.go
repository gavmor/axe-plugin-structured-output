package main_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	protocol "github.com/gavmor/axe-protocol"
	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/gavmor/wasm-microkernel/plugintest"
)

func buildPlugin(t *testing.T) []byte {
	t.Helper()
	wasmPath := filepath.Join(t.TempDir(), "plugin.wasm")

	cmd := exec.Command("go", "build", "-buildmode=c-shared", "-o", wasmPath, ".")
	cmd.Dir = "./"
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile plugin: %v\nOutput: %s", err, string(out))
	}

	data, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestPlugin_ABIConformance(t *testing.T) {
	wasmBytes := buildPlugin(t)
	if err := abi.ValidateABI(wasmBytes, []string{"Metadata", "Execute", "allocate"}, nil); err != nil {
		t.Fatalf("ABI validation failed: %v", err)
	}
}

func TestPlugin_Metadata(t *testing.T) {
	ctx := context.Background()
	wasmBytes := buildPlugin(t)

	h := plugintest.New(t)
	defer h.Close()

	if err := h.Load(ctx, wasmBytes); err != nil {
		t.Fatal(err)
	}

	result, err := h.CallExport(ctx, "Metadata")
	if err != nil {
		t.Fatal(err)
	}

	var def protocol.ToolDefinition
	if err := json.Unmarshal(result, &def); err != nil {
		t.Fatal(err)
	}

	if def.Name != "structured_output_validator" {
		t.Errorf("expected name 'structured_output_validator', got %q", def.Name)
	}
	if def.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestPlugin_Execute_Success(t *testing.T) {
	ctx := context.Background()
	wasmBytes := buildPlugin(t)

	h := plugintest.New(t)
	defer h.Close()

	if err := h.Load(ctx, wasmBytes); err != nil {
		t.Fatal(err)
	}

	call := protocol.ToolCall{
		Name: "structured_output_validator",
		Arguments: map[string]string{
			"provider": "anthropic",
			"format":   "json",
		},
	}
	callBytes, _ := json.Marshal(call)

	offset, err := h.Allocate(ctx, uint32(len(callBytes)))
	if err != nil {
		t.Fatal(err)
	}
	abi.WriteGuestBuffer(ctx, h.Module, offset, callBytes)

	result, err := h.CallExport(ctx, "Execute", uint64(offset), uint64(len(callBytes)))
	if err != nil {
		t.Fatal(err)
	}

	var tr protocol.ToolResult
	if err := json.Unmarshal(result, &tr); err != nil {
		t.Fatal(err)
	}

	if tr.IsError {
		t.Fatalf("expected success, got error: %s", tr.Content)
	}
	if tr.Content != "valid" {
		t.Errorf("expected 'valid', got %q", tr.Content)
	}
}

func TestPlugin_Execute_BedrockFailure(t *testing.T) {
	ctx := context.Background()
	wasmBytes := buildPlugin(t)

	h := plugintest.New(t)
	defer h.Close()

	if err := h.Load(ctx, wasmBytes); err != nil {
		t.Fatal(err)
	}

	call := protocol.ToolCall{
		Name: "structured_output_validator",
		Arguments: map[string]string{
			"provider": "bedrock",
			"format":   "json",
		},
	}
	callBytes, _ := json.Marshal(call)

	offset, err := h.Allocate(ctx, uint32(len(callBytes)))
	if err != nil {
		t.Fatal(err)
	}
	abi.WriteGuestBuffer(ctx, h.Module, offset, callBytes)

	result, err := h.CallExport(ctx, "Execute", uint64(offset), uint64(len(callBytes)))
	if err != nil {
		t.Fatal(err)
	}

	var tr protocol.ToolResult
	if err := json.Unmarshal(result, &tr); err != nil {
		t.Fatal(err)
	}

	if !tr.IsError {
		t.Fatal("expected error for bedrock")
	}
	if tr.Content != `provider "bedrock" does not support structured output` {
		t.Errorf("unexpected error content: %q", tr.Content)
	}
}
