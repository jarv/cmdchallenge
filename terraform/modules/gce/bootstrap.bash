#!/bin/bash

set -ex
exec &> >(tee -a "/var/tmp/bootstrap-$(date +%Y%m%d-%H%M%S).log")

COPY_DIR="/var/tmp"
sudo mkdir -p /var/lib/cloud/bootstrap
sudo chmod 777 /var/lib/cloud/bootstrap

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
  sudo systemctl daemon-reload
  sudo systemctl enable docker-tls-tcp.socket
  sudo systemctl enable docker-cleanup.timer
  sudo systemctl restart docker-cleanup.timer
  sudo systemctl stop docker
  sudo systemctl start docker-tls-tcp.socket
  sudo systemctl start docker
  sudo systemctl stop update-engine
  sudo systemctl disable update-engine
  sudo docker pull registry.gitlab.com/jarv/cmdchallenge/cmd:latest
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
