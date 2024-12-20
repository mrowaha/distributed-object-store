#!/bin/bash

rm -rf data_beta.db

go run . -store=data_beta.db -process=beta -lease=tcp://localhost:5556
