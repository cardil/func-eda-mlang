//go:build linux && arm64

package ffi

import _ "embed"

//go:embed libs/linux_arm64/libeda_core.so
var embeddedLib []byte
