#!/bin/bash

docker build -t mackerel-plugin-scheduler-latency-kvm:latest .
docker run -v $(pwd):/o mackerel-plugin-scheduler-latency-kvm:latest tar zcf /o/mackerel-plugin-scheduler-latency-kvm.tar.gz mackerel-plugin-scheduler-latency-kvm

