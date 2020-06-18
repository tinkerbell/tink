#!/bin/bash

for i in {1..10}; do
	if docker login -u "$TINKERBELL_REGISTRY_USERNAME" -p "$TINKERBELL_REGISTRY_PASSWORD" localhost; then
		break
	fi
	sleep $i
done
docker push localhost/update-data
docker push localhost/overwrite-data
docker push localhost/bash
