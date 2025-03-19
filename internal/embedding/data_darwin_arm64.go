//go:build multiplexer

package embedding

import _ "embed"

//go:embed celestia-app-v3_Darwin_arm64.tar.gz
var v3binaryCompressed []byte

//go:embed celestia-app-v4_Darwin_arm64.tar.gz
var v4binaryCompressed []byte
