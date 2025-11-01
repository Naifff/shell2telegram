#!/bin/bash

export TB_TOKEN="7891013949:AAH5XBnBuIkNcVYVZ0lx8Kd2ZZBm4kKSlTU"

./shell2telegram_cgo \
  --allow-all \
  --enable-db-logging \
  /date 'date' \
  /uptime 'uptime' \
  /test 'echo "Test command at $(date)"'
