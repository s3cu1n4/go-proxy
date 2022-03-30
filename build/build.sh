#!/bin/bash

project_path=$(cd `dirname $0`; pwd)
project_name="${project_path##*/}"

cd $project_path

# macos server
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o ../bin/darwin_server ../server/server.go 

# linux server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ../bin/linux_server ../server/server.go 

# macos client
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o ../bin/darwin_client ../client/main.go 

# windows client
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a  -o ../bin/client.exe ../client/main.go

echo "build successful"


