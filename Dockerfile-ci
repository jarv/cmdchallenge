FROM gitlab/dind
RUN apt-get update && \
    apt-get install -y make python-pip language-pack-en unzip && \
    wget https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz --no-check-certificate && \
    tar -C /usr/local -xzf go1.8.3.linux-amd64.tar.gz && \
    pip install docker==2.0.2 PyYAML==3.12 && \
    curl -o /tmp/tf.zip https://releases.hashicorp.com/terraform/0.11.7/terraform_0.11.7_linux_amd64.zip && \
    cd /usr/local/bin && unzip /tmp/tf.zip
