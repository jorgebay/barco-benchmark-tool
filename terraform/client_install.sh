#! /bin/bash
sudo apt-get update
sudo apt install -y build-essential nghttp2-client
sudo apt-get install -y fio
gcc --version

wget https://go.dev/dl/go1.19.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go*.linux-amd64.tar.gz
sudo sh -c 'echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile'
