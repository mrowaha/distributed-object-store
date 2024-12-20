#!/bin/bash

# Execute the first command to create object named hello
go run . -cmd=1 -object=hello

# Execute the second command to create object named world
go run . -cmd=1 -object=world

# Execute the third commad to delete object named hello
go run . -cmd=2 -object=hello

# Execute the fourth command to delete object named world
go run . -cmd=2 -object=world
