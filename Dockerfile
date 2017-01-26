FROM ubuntu:16.04
RUN apt-get update && apt-get install -y jq bc rename
ADD var.tar.gz /
