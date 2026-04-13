//go:build wasip1

package structured

import (
	"fmt"

	protocol "github.com/gavmor/axe-protocol"
)

// HandleCall validates whether a provider supports structured output.
func HandleCall(call protocol.ToolCall) protocol.ToolResult {
	provider := call.Arguments["provider"]

	supported := false
	switch provider {
	case "anthropic", "openai", "ollama", "google", "gemini":
		supported = true
	}

	if !supported {
		return protocol.ToolResult{
			CallID:  call.ID,
			Content: fmt.Sprintf("provider %q does not support structured output", provider),
			IsError: true,
		}
	}

	return protocol.ToolResult{
		CallID:  call.ID,
		Content: "valid",
	}
}
