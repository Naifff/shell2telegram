#!/bin/bash

# First build with CGO enabled (run this once):
# CGO_ENABLED=1 go build -o shell2telegram_cgo

export TB_TOKEN="7891013949:AAH5XBnBuIkNcVYVZ0lx8Kd2ZZBm4kKSlTU"

# Check if CGO binary exists, if not build it
if [ ! -f "./shell2telegram_cgo" ]; then
  echo "Building shell2telegram with CGO support..."
  CGO_ENABLED=1 go build -o shell2telegram_cgo
fi

./shell2telegram_cgo \
  --allow-all \
  --enable-db-logging \
  /date 'date' \
  /uptime 'uptime' \
  /test 'echo "Test command at $(date)"'
