#!/bin/bash

rm -rf data_alpha.db

if [ "$1" == "daemon" ]; then
    echo "Running in daemon mode..."
    source ./daemon/env/bin/activate
    python daemon -s=data_alpha.db -l=tcp://localhost:5555 -proc=alpha -m="$2"
else
    echo "Running in normal mode..."
    go run . -store=data_alpha.db -process=alpha -lease=tcp://localhost:5555
fi
