package main

import (
	"encoding/json"

	protocol "github.com/gavmor/axe-protocol"
	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/jrswab/axe-plugin-structured-output/internal/structured"
)

func main() {}

//go:wasmexport Metadata
func Metadata() uint64 {
	data, _ := json.Marshal(protocol.ToolDefinition{
		Name:        "structured_output_validator",
		Description: "Validates if a provider supports the requested structured output format.",
		Parameters: map[string]protocol.ToolParameter{
			"provider": {Type: "string", Description: "The provider name (e.g. anthropic, openai)", Required: true},
			"format":   {Type: "string", Description: "The requested format (json or schema)", Required: true},
		},
	})
	return abi.ReturnBytes(data)
}

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(reqBytes []byte) []byte {
		var call protocol.ToolCall
		if err := json.Unmarshal(reqBytes, &call); err != nil {
			res, _ := json.Marshal(protocol.ToolResult{Content: err.Error(), IsError: true})
			return res
		}
		result := structured.HandleCall(call)
		res, _ := json.Marshal(result)
		return res
	})
}

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }
