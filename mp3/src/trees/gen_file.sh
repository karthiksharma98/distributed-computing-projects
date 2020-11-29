#!/bin/bash

echo "This is just a sample line appended to create a big file" > dummy.txt
for i in {1..11}; do
  cat dummy.txt | tee >> dummy.txt
done
