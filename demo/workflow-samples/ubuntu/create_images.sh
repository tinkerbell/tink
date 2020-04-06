#!/bin/sh
  
docker build -t 192.168.1.1/ubuntu:base 00-base/
docker push 192.168.1.1/ubuntu:base
docker build -t 192.168.1.1/disk-wipe:v3 01-disk-wipe/ --build-arg REGISTRY=192.168.1.1
docker push 192.168.1.1/disk-wipe:v3
docker build -t 192.168.1.1/disk-partition:v3 02-disk-partition/ --build-arg REGISTRY=192.168.1.1
docker push 192.168.1.1/disk-partition:v3
docker build -t 192.168.1.1/install-root-fs:v3 03-install-root-fs/ --build-arg REGISTRY=192.168.1.1
docker push 192.168.1.1/install-root-fs:v3
docker build -t 192.168.1.1/install-grub:v3 04-install-grub/ --build-arg REGISTRY=192.168.1.1
docker push 192.168.1.1/install-grub:v3

