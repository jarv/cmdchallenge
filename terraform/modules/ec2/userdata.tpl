#!/bin/bash
# vi: ft=bash
set -xeuf -o pipefail

SECONDS=0
exec &> >(tee -a "/tmp/userdata-$(date +%Y%m%d-%H%M%S).log")

BOOTSTRAP="/usr/local/bin/${bootstrap_fname}"

echo "downloading bootstrap.sh with sha ${bootstrap_sha}"
aws s3 cp ${bootstrap_artifact} $BOOTSTRAP
chmod 755 $BOOTSTRAP

cat <<-CMD >/etc/systemd/system/bootstrap.service
  [Unit]
  Description=boostrap
  After=network-online.target

  [Service]
  ExecStart=$BOOTSTRAP

  [Install]
  WantedBy=multi-user.target
CMD

echo "Restoring cmd backup if it exists"
aws s3 cp ${backup_artifact} /tmp/db.sqlite3.bak.gz || true

systemctl enable bootstrap.service
systemctl restart bootstrap
