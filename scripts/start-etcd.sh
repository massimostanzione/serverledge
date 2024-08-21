#!/bin/sh
docker run -d --rm --name Etcd-server \
    --publish 2379:2379 \
    --publish 2380:2380 \
    --env ALLOW_NONE_AUTHENTICATION=yes \
    --env ETCD_ADVERTISE_CLIENT_URLS=http://192.168.1.245:2379 \
    bitnami/etcd:latest
