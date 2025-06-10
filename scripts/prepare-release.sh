#!/bin/bash
set -e

VERSION="1.0.0"
BINARY_NAME="tips"

echo "Building binaries for macOS..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s -X main.Version=$VERSION" -o ${BINARY_NAME}-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s -X main.Version=$VERSION" -o ${BINARY_NAME}-darwin-arm64

echo "Creating universal binary..."
lipo -create -output ${BINARY_NAME} ${BINARY_NAME}-darwin-amd64 ${BINARY_NAME}-darwin-arm64

echo "Creating tarball..."
tar czf ${BINARY_NAME}-${VERSION}-darwin-universal.tar.gz ${BINARY_NAME} LICENSE README.md

echo "Calculating SHA256..."
shasum -a 256 ${BINARY_NAME}-${VERSION}-darwin-universal.tar.gz

echo "Cleaning up..."
rm ${BINARY_NAME}-darwin-amd64 ${BINARY_NAME}-darwin-arm64 ${BINARY_NAME}

echo "Release artifacts ready!"
echo "Upload ${BINARY_NAME}-${VERSION}-darwin-universal.tar.gz to GitHub releases"