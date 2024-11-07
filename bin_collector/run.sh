#!/bin/ash
set -e

# Load configuration from config.json
ADDRESS=$(jq --raw-output '.address' /data/options.json)
if [ $? -ne 0 ]; then
    echo "Error loading configuration from config.json" >&2
    exit 1
fi

echo "Starting bin-waste-collection with address: $ADDRESS"
/app/bin-waste-collection --address "$ADDRESS"
if [ $? -ne 0 ]; then
    echo "Error running bin-waste-collection" >&2
    exit 1
fi