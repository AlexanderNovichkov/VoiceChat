#!/bin/bash
protoc --python_out=./gen --mypy_out=./gen -I=../proto ../proto/*