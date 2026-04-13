//go:build !wasip1

package structured

import protocol "github.com/gavmor/axe-protocol"

// HandleCall is a no-op on host for test compilation.
func HandleCall(call protocol.ToolCall) protocol.ToolResult {
	return protocol.ToolResult{}
}
