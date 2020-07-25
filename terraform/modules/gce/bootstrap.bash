#!/bin/bash

set -ex
exec &> >(tee -a "/var/tmp/bootstrap-$(date +%Y%m%d-%H%M%S).log")

COPY_DIR="/var/tmp"
BASE_PROM_DIR="/var/lib/docker/prometheus"
NODE_EXPORTER_VERSION="1.0.1"

sudo mkdir -p /var/lib/cloud/bootstrap
sudo chmod 777 /var/lib/cloud/bootstrap
mkdir -p "$BASE_PROM_DIR"

if [[ ! -f "$BASE_PROM_DIR/node_exporter" ]]; then
  curl -L -o /tmp/node-exporter.tar.gz "https://github.com/prometheus/node_exporter/releases/download/v1.0.1/node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz"
  tar -C /tmp -zxf /tmp/node-exporter.tar.gz
  cp /tmp/node_exporter-*/node_exporter "$BASE_PROM_DIR/node_exporter"
  rm -rf /tmp/node_exporter-*
fi

if [[ ! -d "$BASE_PROM_DIR/prometheus" ]]; then
  curl -L -o /tmp/prometheus.tar.gz "https://github.com/prometheus/prometheus/releases/download/v2.20.0/prometheus-2.20.0.linux-amd64.tar.gz"
  tar -C /tmp -zxf /tmp/prometheus.tar.gz
  mv /tmp/prometheus-* "$BASE_PROM_DIR/prometheus"
fi


if [[ -d "$COPY_DIR/docker_cfg_files" ]]; then
  sudo mkdir -p /etc/docker
  sudo mkdir -p /etc/systemd/system/docker.service.d
  sudo cp $COPY_DIR/*cmdchallenge.com/* /etc/docker/
  sudo cp $COPY_DIR/ca.pem /etc/docker/
  sudo cp -r $COPY_DIR/ro_volume /var/
  sudo chmod -v 0400 /etc/docker/server-key.pem
  sudo chmod -v 0444 /etc/docker/ca.pem /etc/docker/server-cert.pem
  sudo cp $COPY_DIR/docker_cfg_files/10-tls-verify.conf /etc/systemd/system/docker.service.d/10-tls-verify.conf
  sudo cp $COPY_DIR/docker_cfg_files/docker-tls-tcp.socket /etc/systemd/system/docker-tls-tcp.socket
  sudo cp $COPY_DIR/docker_cfg_files/docker-cleanup.timer /etc/systemd/system/docker-cleanup.timer
  sudo cp $COPY_DIR/docker_cfg_files/docker-cleanup.service /etc/systemd/system/docker-cleanup.service
  sudo cp $COPY_DIR/docker_cfg_files/prometheus.service /etc/systemd/system/prometheus.service
  sudo cp $COPY_DIR/docker_cfg_files/node-exporter.service /etc/systemd/system/node-exporter.service
  sudo cp $COPY_DIR/prometheus.yml "$BASE_PROM_DIR/prometheus/prometheus.yml"
  sudo systemctl daemon-reload
  sudo systemctl enable docker-tls-tcp.socket
  sudo systemctl enable docker-cleanup.timer
  sudo systemctl restart docker-cleanup.timer
  sudo systemctl enable prometheus.service
  sudo systemctl enable node-exporter.service
  sudo systemctl restart prometheus.service
  sudo systemctl restart node-exporter.service
  sudo systemctl stop docker
  sudo systemctl start docker-tls-tcp.socket
  sudo systemctl start docker
  sudo systemctl stop update-engine
  sudo systemctl disable update-engine
  sudo docker pull registry.gitlab.com/jarv/cmdchallenge/cmd:latest
  sudo docker pull registry.gitlab.com/jarv/cmdchallenge/cmd-no-bin:latest
fi

if [[ ! -f /var/swapfile ]]; then
  # Enable swap
  sudo sysctl vm.disk_based_swap=1
  sudo fallocate -l 1G /var/swapfile
  sudo chmod 600 /var/swapfile
  sudo mkswap /var/swapfile
  sudo swapon /var/swapfile
fi

# Allow connections from 2376 for docker tls
sudo iptables -w -A INPUT -p tcp --dport 2376 -j ACCEPT
sudo iptables -w -A INPUT -p tcp --dport 9090 -j ACCEPT
