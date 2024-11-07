#!/bin/bash

# Load configuration from options.json
ADDRESS=$(jq -r '.address' /data/options.json)

# Export environment variable for the Go application
export ADDRESS

# Ensure the app is executable (not usually needed if set during build)
chmod +x app

# Run the Go application
./app
