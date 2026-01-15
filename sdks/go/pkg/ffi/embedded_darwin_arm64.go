//go:build darwin && arm64

package ffi

import _ "embed"

//go:embed libs/darwin_arm64/libeda_core.dylib
var embeddedLib []byte
