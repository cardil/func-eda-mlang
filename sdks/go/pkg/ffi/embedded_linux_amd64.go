//go:build linux && amd64

package ffi

import _ "embed"

//go:embed libs/linux_amd64/libeda_core.so
var embeddedLib []byte
