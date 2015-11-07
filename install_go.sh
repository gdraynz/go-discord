#!/bin/bash

if [ -d "$CACHE_GODEB" ]; then
    sudo dpkg -i /home/ubuntu/.godeb/go_1.5.1-godeb1_amd64.deb
else
    mkdir $CACHE_GODEB
    cd $CACHE_GODEB && godeb install 1.5.1
fi
