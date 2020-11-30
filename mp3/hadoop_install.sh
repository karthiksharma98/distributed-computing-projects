#!/bin/bash

# Make sure to run git config --global credential.helper store once to save username/pass
 
for val in fa20-cs425-g13-0{1..9}.cs.illinois.edu; do
       ssh $1@$val -t "wget https://downloads.apache.org/hadoop/common/hadoop-2.9.2/hadoop-2.9.2.tar.gz && tar -xvzf hadoop-2.9.2.tar.gz && mv hadoop-2.9.2 /home/ksharma/hadoop"
done

ssh $1@fa20-cs425-g13-10.cs.illinois.edu -t "wget https://downloads.apache.org/hadoop/common/hadoop-2.9.2/hadoop-2.9.2.tar.gz && tar -xvzf hadoop-2.9.2.tar.gz && mv hadoop-2.9.2 /home/ksharma/hadoop"
