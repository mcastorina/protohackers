#!/bin/bash

HOST=localhost
PORT=1337

function camera1 {
    printf '\x80\x00\x7b\x00\x08\x00\x3c'
}

function camera2 {
    printf '\x80\x00\x7b\x00\x09\x00\x3c'
}

function plate1 {
    printf '\x20\x04\x55\x4e\x31\x58\x00\x00\x00\x00'
}

function plate2 {
    printf '\x20\x04\x55\x4e\x31\x58\x00\x00\x00\x2d'
}

# Hexadecimal:
# <-- 80 00 7b 00 08 00 3c
# <-- 20 04 55 4e 31 58 00 00 00 00

# Decoded:
# <-- IAmCamera{road: 123, mile: 8, limit: 60}
# <-- Plate{plate: "UN1X", timestamp: 0}

(camera1; plate1) | nc "$HOST" "$PORT" &

# Client 2: camera at mile 9

# Hexadecimal:
# <-- 80 00 7b 00 09 00 3c
# <-- 20 04 55 4e 31 58 00 00 00 2d

# Decoded:
# <-- IAmCamera{road: 123, mile: 9, limit: 60}
# <-- Plate{plate: "UN1X", timestamp: 45}

(camera2; plate2) | nc "$HOST" "$PORT" &
