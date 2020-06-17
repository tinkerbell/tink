#!/bin/bash

while ! docker login -u username -p password localhost; do
	sleep 1
done
docker push localhost/update-data
docker push localhost/overwrite-data
docker push localhost/bash
