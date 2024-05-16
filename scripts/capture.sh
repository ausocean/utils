#!/bin/bash
# Capture some video using raspivid at 640x480 resolution and 25 fps and either write to the screen or a file.
Screen=""
Output="video.h264"
echo "$1"
if [ -n "$1" ]; then
    Output="$1"
fi
if [ -n "$2" ]; then
    Screen="-p 0,0,640,480"
fi
echo Running raspivid for 30s, writing to "$Output"
if [ -n "$Screen" ]; then
    raspivid -t 30000 -w 640 -h 480 -fps 25 -b 1200000 -p "$Screen" -o "$Output"
else
    raspivid -t 30000 -w 640 -h 480 -fps 25 -b 1200000 -o "$Output"
fi
