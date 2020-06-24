FROM ubuntu:16.04
RUN apt-get update && apt-get install -y jq bc rename bsdmainutils man && rm -f /etc/bash.bashrc && rm -rf /etc/bash_completion.d && rm -f /root/.bashrc
ADD var.tar.gz /
