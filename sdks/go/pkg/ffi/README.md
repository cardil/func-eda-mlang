# FFI Package - Embedded Multi-Platform Libraries

Self-contained FFI bindings to Rust `eda-core` using embedded libraries and `purego`.

## Features

- ✅ No external library dependencies - libraries embedded at compile time
- ✅ No CGO for FFI calls - uses `purego` for dynamic loading
- ✅ Cross-platform - Linux, macOS, Windows (amd64/arm64)
- ✅ No `LD_LIBRARY_PATH` needed - works out of the box

## Structure

```
pkg/ffi/
├── ffi.go                      # Core interface
├── loader.go                   # Dynamic loader (purego)
├── embedded_*.go               # Platform-specific embeds
└── libs/
    ├── linux_amd64/libeda_core.so
    ├── linux_arm64/libeda_core.so
    ├── darwin_amd64/libeda_core.dylib
    ├── darwin_arm64/libeda_core.dylib
    └── windows_amd64/eda_core.dll
```

## Building

```bash
# Build for current platform
make sdk-go-embed-libs
make sdk-go-build

# Cross-compile all platforms
make core-build-ffi-all
make sdk-go-embed-libs-all
```

## Usage

```go
core, err := ffi.NewCore()
if err != nil {
    panic(err)
}
defer core.Close()

config, _ := core.GetKafkaConfig()
```

## How It Works

1. Libraries built with Rust and copied to `libs/`
2. Platform-specific `//go:embed` includes correct library
3. At runtime, library extracted to temp file and loaded with `purego`
4. C functions registered and called without CGO

## Testing

```bash
make example-go-ffi
ldd sdks/go/examples/ffi-example/ffi-consumer | grep eda  # No dependency!
make run-go-ffi  # Works without LD_LIBRARY_PATH
```
