#!/bin/bash

docker pull bash
docker tag bash:latest localhost/bash
docker build -t localhost/update-data actions/update_data/
docker build -t localhost/overwrite-data actions/overwrite_data/
