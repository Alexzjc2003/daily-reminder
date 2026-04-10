#!/usr/bin/env bash
set -euo pipefail

GO111MODULE=off go build -o ./build/dr .
