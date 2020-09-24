#!/bin/bash

# Make sure to run git config --global credential.helper store once to save username/pass
 
for val in fa20-cs425-g13-0{1..9}.cs.illinois.edu; do
   ssh $1@$val -t 'cd cs425_mps && git pull && cd src && go build -o main && ls -l'
done

ssh $1@fa20-cs425-g13-10.cs.illinois.edu -t 'cd cs425_mps && git pull && cd src && go build -o main && ls -l'