//go:build windows && amd64

package ffi

import _ "embed"

//go:embed libs/windows_amd64/eda_core.dll
var embeddedLib []byte
