#!/bin/bash

# Loop 30000 times
for i in {1..30000}
do
  echo "Executing command iteration $i"
 ../bin/serverledge-cli invoke -H 192.168.1.245 -P 1324 -f func -p "a:2" -p "b:3"
done
