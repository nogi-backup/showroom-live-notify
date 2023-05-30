#!/bin/bash

cd showroom &&
    GOOS=linux GOARCH=amd64 go build -o showroom . &&
    zip showroom.zip showroom
