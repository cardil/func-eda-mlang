//go:build darwin && amd64

package ffi

import _ "embed"

//go:embed libs/darwin_amd64/libeda_core.dylib
var embeddedLib []byte
