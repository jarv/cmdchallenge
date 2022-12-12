FROM ubuntu:22.04
COPY serve /usr/local/bin/runcmd
COPY oops /usr/local/bin/oops-this-will-delete-bin-dirs
RUN apt-get update && \
      apt-get install -y jq bc rename bsdmainutils man file && \
      rm -f /etc/bash.bashrc && rm -rf /etc/bash_completion.d && \
      rm -f /root/.bashrc && \
      ln -s /ro_volume/ch /usr/local/bin/ch
ADD var/ /var/
