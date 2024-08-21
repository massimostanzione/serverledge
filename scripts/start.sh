#!/bin/sh
sudo docker run -d --rm --name Etcd-server \
    --publish 2379:2379 \
    --publish 2380:2380 \
    --env ALLOW_NONE_AUTHENTICATION=yes \
    --env ETCD_ADVERTISE_CLIENT_URLS=http://192.168.1.245:2379 \
    bitnami/etcd:latest

#!/bin/sh
sudo docker run -d --rm -p 8086:8086 --name InfluxDb     -e DOCKER_INFLUXDB_INIT_MODE=setup \
            -e DOCKER_INFLUXDB_INIT_USERNAME=user \
            -e DOCKER_INFLUXDB_INIT_PASSWORD=password \
            -e DOCKER_INFLUXDB_INIT_ORG=serverledge \
            -e DOCKER_INFLUXDB_INIT_BUCKET=stats \
      -e DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=serverledge \
      influxdb
