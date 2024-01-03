# Basic Makefile for a Go project

# Binary name
BINARY_NAME=./bin/1brc.bin

# Build the project
all: fresh

# Fresh command (clean, build, and run)
fresh: clean build run

# Build command
build: 
	CGO_ENABLED=0 go build -o $(BINARY_NAME) -ldflags="-s -w" -v ./

# Run the project
run:
	$(BINARY_NAME) --file=measurements.txt

# Clean build files
clean: 
	go clean
	rm -f $(BINARY_NAME)

# Make sure these targets are executed as commands
.PHONY: all build run clean
