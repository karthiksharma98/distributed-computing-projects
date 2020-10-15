#!/bin/bash

# Make sure to run git config --global credential.helper store once to save username/pass
 
for val in fa20-cs425-g13-0{1..9}.cs.illinois.edu; do
   if ssh $1@$val '[ ! -d ~/cs425_mps ]'
   then
       ssh $1@$val -t 'git clone https://gitlab.engr.illinois.edu/ksharma/cs425_mps.git'
   else 
       ssh $1@$val -t "cd cs425_mps && git fetch && git checkout $2 && git pull && cd mp2/src && rm -f machine.log.txt && bash ../build.sh && ls -l"
   fi
done

ssh $1@fa20-cs425-g13-10.cs.illinois.edu -t "cd cs425_mps && git fetch && git checkout $2 && cd mp2/src && rm -f machine.log.txt && bash ../build.sh && ls -l"
