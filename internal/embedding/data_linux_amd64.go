//go:build multiplexer

package embedding

import _ "embed"

//go:embed celestia-app-v3_Linux_x86_64.tar.gz
var v3binaryCompressed []byte

var v4binaryCompressed []byte
