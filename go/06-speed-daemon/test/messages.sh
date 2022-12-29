#!/bin/bash

HOST=localhost
PORT=1337

function camera {
    printf '\x80\x00\x42\x00\x64\x00\x3c'
}

function plate {
    printf '\x20\x04\x55\x4e\x31\x58\x00\x00\x03\xe8'
}

(camera; plate) | nc "$HOST" "$PORT" | xxd
