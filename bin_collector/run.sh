#!/bin/bash

# Load configuration from options.json
if [ -f "/data/options.json" ]; then
    ADDRESS=$(jq --raw-output '.address' /data/options.json)
    if [ $? -eq 0 ]; then
        # Export environment variable for the Go application
        export ADDRESS
        
        # Run the Go application
        ./bin-waste-collection
    else
        echo "Error: Failed to extract address from options.json" >&2
        exit 1
    fi
else
    echo "Error: options.json file not found" >&2
    exit 1
fi