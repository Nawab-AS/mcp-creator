package utils

import (
	"mcp-creator/utils/files"
	"mcp-creator/utils/network"
)

// aggregator for all utils

type Utils struct {
	Files   *files.Files
	Network *network.Network
}

func (u *Utils) Hash(input string) string {
	return files.Hash(input)
}
