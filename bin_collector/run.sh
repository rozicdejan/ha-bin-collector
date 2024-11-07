#!/bin/sh

# Check if an ADDRESS environment variable is set, if not, use a default value
if [ -z "$ADDRESS" ]; then
  export ADDRESS="začret 69"
  echo "No ADDRESS environment variable set, using default: $ADDRESS"
else
  echo "Using ADDRESS environment variable: $ADDRESS"
fi

# Start the Go application
./main
