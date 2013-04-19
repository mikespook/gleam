#!/bin/bash

check_base_env() {
if [ ! -n "$SERVICE_BASE" ]; then
    if [ -n "$1" ]; then
        SERVICE_BASE=$1
    else
        SERVICE_BASE=$HOME
    fi
fi
}
