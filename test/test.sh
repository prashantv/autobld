#!/bin/sh
STARTED=`date +%s`
echo "sh: go run started at $STARTED"
go run go/test.go
echo "sh: go run returned after starting at $STARTED"
