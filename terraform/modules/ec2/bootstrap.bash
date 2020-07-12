#!/bin/bash
DIR=/home/core
set -ex
exec &> >(tee -a "/tmp/bootstrap.log")
sudo mkdir -p /etc/docker
sudo mkdir -p /etc/systemd/system/docker.service.d
sudo mv $DIR/runcmd/private/* /etc/docker/
sudo mv $DIR/runcmd/ro_volume /var/
sudo chmod -v 0400 /etc/docker/server-key.pem
sudo chmod -v 0444 /etc/docker/ca.pem /etc/docker/server-cert.pem
sudo mv $DIR/runcmd/docker_cfg_files/10-tls-verify.conf /etc/systemd/system/docker.service.d/10-tls-verify.conf
sudo mv $DIR/runcmd/docker_cfg_files/docker-tls-tcp.socket /etc/systemd/system/docker-tls-tcp.socket
sudo mv $DIR/runcmd/docker_cfg_files/docker-cleanup.timer /etc/systemd/system/docker-cleanup.timer
sudo mv $DIR/runcmd/docker_cfg_files/docker-cleanup.service /etc/systemd/system/docker-cleanup.service
sudo mv $DIR/runcmd/docker_cfg_files/swap.service /etc/systemd/system/swap.service
sudo systemctl daemon-reload
sudo systemctl enable --now /etc/systemd/system/swap.service
sudo systemctl enable docker-tls-tcp.socket
sudo systemctl enable docker-cleanup.timer
sudo systemctl restart docker-cleanup.timer
sudo systemctl stop docker
sudo systemctl start docker-tls-tcp.socket
sudo systemctl start docker
sudo systemctl stop update-engine
sudo systemctl disable update-engine

docker pull registry.gitlab.com/jarv/cmdchallenge/cmd:testing
docker pull registry.gitlab.com/jarv/cmdchallenge/cmd:latest
