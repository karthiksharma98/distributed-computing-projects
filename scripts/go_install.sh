#!/bin/bash

wget https://golang.org/dl/go1.15.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.linux-amd64.tar.gz

echo 'export GOPATH="$HOME/Go"i' >> ~/.bashrc # or any directory to put your Go code
echo 'export PATH="$PATH:/usr/local/go/bin:$GOPATH/bin"' >> ~/.bashrc

