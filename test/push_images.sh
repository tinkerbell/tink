#!/bin/bash

for i in {1..10}
do
    docker login -u username -p password localhost
    if [ $? -eq 0 ]; then
        break
    fi
    sleep 1
done
docker push localhost/action1
docker push localhost/action2
docker push localhost/action3
