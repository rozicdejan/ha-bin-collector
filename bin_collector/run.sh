#!/bin/ash
set -e

# Load configuration from config.json
ADDRESS=$(jq --raw-output '.address' /data/config.json)

# Run the Go application with the address argument
/app/bin-waste-collection --address "$ADDRESS"