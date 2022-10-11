#! /bin/bash
sudo apt-get update
sudo apt install -y build-essential nghttp2-client
sudo apt-get install -y fio
gcc --version

wget https://go.dev/dl/go1.19.2.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go*.linux-arm64.tar.gz
sudo sh -c 'echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile'

sudo mkdir /data
sudo mkfs -t xfs /dev/nvme1n1
sudo mount /dev/nvme1n1 /data
sudo chmod 777 /data
sudo mkdir /data/barco-data
sudo chmod 777 /data/barco-data
