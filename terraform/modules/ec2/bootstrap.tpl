#!/bin/bash
# vi: ft=bash
set -xeuf -o pipefail

SECONDS=0
NODE_EXPORTER_VERSION="1.2.2"
PROMETHEUS_VERSION="2.30.0"
BASE_PROM_DIR="/opt/prometheus"
PROM_DATA_DIR="/var/opt/prometheus"
CMD_DATA_DIR="/var/opt/cmd"
SQLITE_DB_FILE="$CMD_DATA_DIR/db.sqlite3"
RO_VOLUME_DIR="/var/opt/ro_volume"
CMD_USER="cmd"

prep() {
  yum update -y
  amazon-linux-extras install docker vim -y
  yum install htop -y
  yum erase amazon-ssm-agent -y
  usermod -a -G docker ec2-user
}

configSwap() {
  [[ -f /var/swapfile ]] && return

  sudo fallocate -l 1G /var/swapfile
  sudo chmod 600 /var/swapfile
  sudo mkswap /var/swapfile
  sudo swapon /var/swapfile
}

configNodeExporter() {
  node_exporter_dir="node_exporter-$NODE_EXPORTER_VERSION.linux-amd64"
  node_exporter_fname="$node_exporter_dir.tar.gz"
  [[ -f /tmp/$node_exporter_fname ]] && return

  mkdir -p "$BASE_PROM_DIR"
  id -u node_exporter &>/dev/null || useradd node_exporter

  curl -L -o "/tmp/$node_exporter_fname" "https://github.com/prometheus/node_exporter/releases/download/v$NODE_EXPORTER_VERSION/$node_exporter_fname"
  tar zxf "/tmp/$node_exporter_fname" -C /tmp
  rm -rf "$BASE_PROM_DIR/node_exporter"
  mv /tmp/$node_exporter_dir "$BASE_PROM_DIR/node_exporter"
  chown -R node_exporter "$BASE_PROM_DIR/node_exporter"

  cat <<-SERVICE >/etc/systemd/system/node_exporter.service
  [Unit]
  Description=Prometheus Server
  Documentation=https://node_exporter.io/docs/introduction/overview/
  After=network-online.target

  [Service]
  User=node_exporter
  Restart=on-failure

  ExecStart=$BASE_PROM_DIR/node_exporter/node_exporter

  [Install]
  WantedBy=multi-user.target
	SERVICE

  chmod 755 /etc/systemd/system/node_exporter.service
  systemctl enable node_exporter.service
  systemctl restart node_exporter
}

configPrometheus() {
  prometheus_dir="prometheus-$PROMETHEUS_VERSION.linux-amd64"
  prometheus_fname="$prometheus_dir.tar.gz"

  [[ -f /tmp/$prometheus_fname ]] && return

  mkdir -p "$BASE_PROM_DIR"
  mkdir -p "$PROM_DATA_DIR"
  id -u prometheus &>/dev/null || useradd prometheus

  curl -L -o "/tmp/$prometheus_fname" "https://github.com/prometheus/prometheus/releases/download/v$PROMETHEUS_VERSION/$prometheus_fname"
  tar zxf "/tmp/$prometheus_fname" -C /tmp
  rm -rf "$BASE_PROM_DIR/prometheus"
  mv /tmp/$prometheus_dir "$BASE_PROM_DIR/prometheus"
  cat <<-PROM >$BASE_PROM_DIR/prometheus/prometheus.yml
		global:
		  scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
		  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
		  # scrape_timeout is set to the global default (10s).
		scrape_configs:
		  - job_name: 'node_exporter'
		    static_configs:
		    - targets: ['localhost:9100']
		  - job_name: 'prometheus'
		    static_configs:
		    - targets: ['localhost:9090']
		  - job_name: 'docker'
		    static_configs:
		    - targets: ['localhost:9323']
		  - job_name: 'cmd'
		    metrics_path: /metrics
		    static_configs:
		    - targets: ['localhost:8181']
	PROM

  cat <<-'WEBCFG' >$BASE_PROM_DIR/prometheus/web.yml
		basic_auth_users:
		  # cmd/cmd
		  cmd: $2a$12$PdF9KPI4LYzAVB6a.l0rYe28ZggWPnlT9y0/uSaa2tXrLzM90luIK
	WEBCFG

  chown -R prometheus "$BASE_PROM_DIR"
  chown -R prometheus "$PROM_DATA_DIR"

  cat <<-SERVICE >/etc/systemd/system/prometheus.service
		[Unit]
		Description=Prometheus Server
		Documentation=https://prometheus.io/docs/introduction/overview/
		After=network-online.target

		[Service]
		User=prometheus
		Restart=on-failure

		ExecStart=$BASE_PROM_DIR/prometheus/prometheus \
		  --config.file=$BASE_PROM_DIR/prometheus/prometheus.yml \
		  --storage.tsdb.path=$PROM_DATA_DIR \
      --web.config.file=$BASE_PROM_DIR/prometheus/web.yml

		[Install]
		WantedBy=multi-user.target
	SERVICE

  chmod 755 /etc/systemd/system/prometheus.service
  systemctl enable prometheus.service
  systemctl restart prometheus.service
}

restoreDB() {
  db_backup_file="/tmp/db.sqlite3.bak.gz"
  ! [[ -r "$db_backup_file" ]] && return

  gzip -d "$db_backup_file"
  echo '.restore /tmp/db.sqlite3.bak' | sqlite3 "$SQLITE_DB_FILE"
  rm -f /tmp/db.sqlite3.bak
  chown $CMD_USER:$CMD_USER "$SQLITE_DB_FILE"
}

configCmd() {
  rm -rf "$RO_VOLUME_DIR"
  rm -f /usr/local/bin/serve

  id -u $CMD_USER &>/dev/null || useradd $CMD_USER -G docker
  mkdir -p "$CMD_DATA_DIR"
  chown $CMD_USER "$CMD_DATA_DIR"

  restoreDB

  aws s3 cp ${ro_volume_artifact} /tmp/ro_volume.tar.gz
  tar xzf /tmp/ro_volume.tar.gz -C /var/opt
  rm -f /tmp/ro_volume.tar.gz
  aws s3 cp ${serve_artifact} /usr/local/bin/serve
  chmod 755 /usr/local/bin/serve
  cat <<-CMD >/etc/systemd/system/cmd.service
		[Unit]
		Description=cmd
		After=network-online.target

		[Service]
		Environment="SQLITE_DB_FILE=$SQLITE_DB_FILE"
		Environment="RO_VOLUME_DIR=$RO_VOLUME_DIR"
		Environment="CMD_IMAGE_TAG=${cmd_image_tag}"
		PassEnvironment=SQLITE_DB_FILE RO_VOLUME_DIR SERVERPORT CMD_IMAGE_TAG
		User=$CMD_USER
		Restart=on-failure

		ExecStart=/usr/local/bin/serve -addr :8181 ${cmd_extra_opts}

		[Install]
		WantedBy=multi-user.target
	CMD

  chmod 755 /etc/systemd/system/cmd.service
  systemctl enable cmd.service
  systemctl restart cmd.service
}

pullImages() {
  docker pull "registry.gitlab.com/jarv/cmdchallenge/cmd:${cmd_image_tag}"
  docker pull "registry.gitlab.com/jarv/cmdchallenge/cmd-no-bin:${cmd_image_tag}"
}

configDocker() {
  [[ -r /etc/docker/daemon.json ]] && return

  cat <<-DAEMON >/etc/docker/daemon.json
		{
		  "live-restore": true,
		  "storage-driver": "overlay2",
		  "mtu": 1460,
		  "metrics-addr" : "127.0.0.1:9323",
		  "experimental" : true
		}
	DAEMON

  systemctl enable docker.service
  systemctl restart docker
}

configJanitor() {
  [[ -r /usr/local/bin/janitor.sh ]] && return

  cat <<-'JANITOR' >/usr/local/bin/janitor.sh
		#!/bin/bash
		set -euf
		docker container prune -f
		docker ps | grep  " hour" | awk "{print \$1}" | xargs -r docker rm -f
	JANITOR

  chmod 755 /usr/local/bin/janitor.sh

  cat <<-CMD >/etc/systemd/system/janitor.service
		[Unit]
		Description=Cleans up stale docker containers
		After=network-online.target

		[Service]
		Type=oneshot
		ExecStart=/usr/local/bin/janitor.sh
		[Install]
		WantedBy=multi-user.target
	CMD

  cat <<-TIMER >/etc/systemd/system/janitor.timer
		[Unit]
		Description=Run janitor.service every 10 minutes

		[Timer]
		OnCalendar=*:0/10
		Persistent=true

		[Install]
		WantedBy=multi-user.target
	TIMER

  systemctl enable janitor.service
  systemctl enable janitor.timer
  systemctl restart janitor.service
  systemctl restart janitor.timer
}

configBackup() {
  cat <<-'BACKUP' >/usr/local/bin/backup.sh
		#!/bin/bash
		set -euf -o pipefail
		db="/var/opt/cmd/db.sqlite3"
		if ! [[ -r "$db" ]]; then
		  echo "Unable to read $db"
		  exit 1
		fi

		backup_fname="$(mktemp).sq3.bak"
		sqlite3 "$db" ".backup $backup_fname"
		gzip "$backup_fname"
		aws s3 cp "$backup_fname.gz" ${backup_artifact}
		rm -f "$backup_fname.gz"
	BACKUP
  chmod 755 /usr/local/bin/backup.sh

  cat <<-CMD >/etc/systemd/system/backup.service
		[Unit]
		Description=backup
		After=network-online.target

		[Service]
		ExecStart=/usr/local/bin/backup.sh

		[Install]
		WantedBy=multi-user.target
	CMD

  cat <<-TIMER >/etc/systemd/system/backup.timer
		[Unit]
		Description=Runs backup.service every 15 minutes

		[Timer]
		OnCalendar=*:0/15
		Persistent=true

		[Install]
		WantedBy=multi-user.target
	TIMER

  systemctl enable backup.service
  systemctl enable backup.timer
  systemctl restart backup.service
  systemctl restart backup.timer
}

# Main

prep
configSwap
configDocker
pullImages
configJanitor
configPrometheus
configNodeExporter
configCmd
configBackup
yum clean all
rm -rf /var/cache/yum/*

duration=$SECONDS
echo "$(date -u): Bootstrap finished in $((duration / 60)) minutes and $((duration % 60)) seconds"
