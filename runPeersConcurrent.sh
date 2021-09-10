#!/bin/bash
n=50

echo "started initialization of ${n} peers"

command="go run . -c config/config1.ini"

for ((i = 2; i <= ${n}; ++i)); do
command="${command} & go run . -c config/config${i}.ini"


done
#echo ${command}
eval $command 
