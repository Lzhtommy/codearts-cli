// codearts-cli — Huawei Cloud CodeArts CLI (Go implementation).
//
// A lightweight CLI wrapping the CodeArts SDK for common CI/CD workflows
// (currently the CodeArtsPipeline RunPipeline API).
package main

import (
	"os"

	"github.com/Lzhtommy/codearts-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
