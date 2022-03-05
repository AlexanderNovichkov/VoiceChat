#!/bin/bash
mkdir gen
protoc --go_out=. -I=../proto ../proto/*