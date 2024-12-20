#!/bin/bash

rm -rf data_charlie.db

if [ "$1" == "daemon" ]; then
    echo "Running in daemon mode..."
    source ./daemon/env/bin/activate
    python daemon -s=data_charlie.db -l=tcp://localhost:5557 -proc=charlie
else
    echo "Running in normal mode..."
    go run . -store=data_charlie.db -process=charlie -lease=tcp://localhost:5557
fi
