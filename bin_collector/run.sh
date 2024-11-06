#!/bin/bash

# Load configuration from config.json
ADDRESS=$(jq -r '.address' /data/options.json)

# Export environment variable for the Go application
export ADDRESS

# Run the Go application
./app
