#!/bin/bash

script_abspath="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

annotate() {
    local src="$script_abspath/$1"
    local dst="$script_abspath/tmp/$2"
    local txt="$3"
    convert "$src" -fill '#ffffff' -gravity Northwest -pointsize 30 -annotate +5+5 "$txt" "$dst"
}

mkdir -p "$script_abspath/tmp"
annotate kodim23-01-jillions.png kodim23-0001.png "Jillions of colours"
annotate kodim23-02-256.png kodim23-0002.png "256 colours"
annotate kodim23-03-128.png kodim23-0003.png "128 colours"
annotate kodim23-04-64.png kodim23-0004.png "64 colours"
annotate kodim23-05-32.png kodim23-0005.png "32 colours"
annotate kodim23-06-16.png kodim23-0006.png "16 colours"
annotate kodim23-07-8.png kodim23-0007.png "8 colours"
annotate kodim23-08-4.png kodim23-0008.png "4 colours"
annotate kodim23-09-2.png kodim23-0009.png "2 colours"

ffmpeg \
    -y -framerate 0.5 \
    -f image2 \
    -i "$script_abspath/tmp/kodim23-%0004d.png" \
    -plays 10 \
    "$script_abspath/kodim23.apng"

