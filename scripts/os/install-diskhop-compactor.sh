#!/bin/bash

# Set the working directory to the script's directory
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

# Define the Go binary directory
GO_BINARY_DIR="$SCRIPT_DIR/../go/dopcompactor"

# Define the binary name and target installation path
BINARY_NAME="dopcompactor"

INSTALL_PATH="$GOPATH/bin/$BINARY_NAME"

# Check if Go is installed
if ! command -v go &>/dev/null; then
  echo "Go is not installed. Please install Go and try again."
  exit 1
fi

# Navigate to the Go binary directory
cd "$GO_BINARY_DIR" || {
  echo "Failed to navigate to $GO_BINARY_DIR"
  exit 1
}

# Compile the Go binary
echo "Compiling $BINARY_NAME..."
go build -o "$BINARY_NAME" "./main.go"

if [ $? -ne 0 ]; then
  echo "Compilation failed. Please check your Go code."
  exit 1
fi

# Move the compiled binary to the Go binary path
echo "Installing $BINARY_NAME to $INSTALL_PATH..."
sudo mv "$BINARY_NAME" "$INSTALL_PATH"

if [ $? -eq 0 ]; then
  echo "$BINARY_NAME installed successfully at $INSTALL_PATH"
else
  echo "Failed to install $BINARY_NAME. Please check your permissions."
  exit 1
fi

# Make sure the binary is executable
sudo chmod +x "$INSTALL_PATH"

# Verify installation
echo "Verifying installation..."
if command -v "$BINARY_NAME" &>/dev/null; then
  echo "$BINARY_NAME is installed and ready to use."
else
  echo "Installation verification failed. $BINARY_NAME is not found in the system PATH."
  exit 1
fi
