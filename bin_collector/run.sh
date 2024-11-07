#!/usr/bin/env bash
set -e

# Load configuration from options.json
ADDRESS=$(jq --raw-output '.address' /data/options.json)

# Run the Go application with the address argument
/app/bin-waste-collection --address "$ADDRESS"