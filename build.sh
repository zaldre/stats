#!/bin/bash
set -e
VERSION=$(grep 'const VERSION' cmd/stats/main.go | sed -E 's/.*"([^"]+)".*/\1/')
BIN_NAME="stats"

# Platforms you want to build for
targets=(
"linux/amd64"
"darwin/arm64"
"darwin/amd64"
"windows/amd64"
)

for target in "${targets[@]}"; do
    GOOS=${target%/*}
    GOARCH=${target#*/}
    output_dir="bin/${GOOS}/${GOARCH}"
    output_file="${BIN_NAME}-v${VERSION}"
    
    # Windows binaries need .exe extension
    if [ "$GOOS" == "windows" ]; then
        output_file="${output_file}.exe"
    fi
    
    mkdir -p "$output_dir"
    
    echo "Building for $GOOS/$GOARCH -> $output_dir/$output_file"
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-s -w -X main.version=$VERSION" \
        -o "${output_dir}/${output_file}" \
        ./cmd/stats
    
    # Compress with UPX (skip macOS as UPX doesn't support Apple Silicon)
    if [ "$GOOS" != "darwin" ] && command -v upx &> /dev/null; then
        echo "Compressing ${output_file} with UPX..."
        upx --best --lzma "${output_dir}/${output_file}"
    fi
done