//go:build multiplexer

package embedding

import _ "embed"

//go:embed celestia-app-v3-pr4430_Linux_arm64.tar.gz
var v3binaryCompressed []byte
