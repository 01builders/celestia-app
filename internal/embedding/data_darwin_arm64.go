package embedding

import _ "embed"

//go:embed celestia-app-v3_Darwin_arm64.tar.gz
var v3binaryCompressed []byte

var v4binaryCompressed []byte
