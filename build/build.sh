#!/bin/bash

project_path=$(cd `dirname $0`; pwd)
project_name="${project_path##*/}"

cd $project_path

# macos server
go build -a ../server/server.go 

# linux server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o linux_server ../server/server.go 

# macos client
go build -a  -o client ../client/main.go 

# windows client
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a  -o client.exe ../client/main.go




echo "build successful"


