#!/bin/bash

: '
  In this script we are going to create two objects
  but we are going to delete the second object
  then it is advised to used sqlite3 exec to valdiate results
'

go run . -cmd=1 -object=me -data=rowaha
go run . -cmd=1 -object=hello -data=world
echo [order_2] checkpoint: attempted create objects
go run . -cmd=2 -object=me
