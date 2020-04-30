#!/bin/bash

set -ex
exec &> >(tee -a "/tmp/bootstrap-cloud.log")
mkdir -p /var/lib/cloud/bootstrap
chmod 777 /var/lib/cloud/bootstrap

# Enable swap
sysctl vm.disk_based_swap=1
fallocate -l 1G /var/swapfile
chmod 600 /var/swapfile
mkswap /var/swapfile
swapon /var/swapfile

# Allow connections from 2376 for docker tls
iptables -w -A INPUT -p tcp --dport 2376 -j ACCEPT
