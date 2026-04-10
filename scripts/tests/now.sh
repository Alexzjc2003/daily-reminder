#!/usr/bin/env bash
set -euo pipefail

./build/dr remember -r $(tm -f "%H:%M:%S") $(tm -f "%Y/%m/%d")
./build/dr status