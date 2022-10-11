git clone https://github.com/jorgebay/barco-benchmark-tool.git
cd barco-benchmark-tool && go build .

./barco-benchmark-tool -u http://10.0.0.100:9251/v1/topic/a-topic/messages,http://10.0.0.101:9251/v1/topic/a-topic/messages,http://10.0.0.102:9251/v1/topic/a-topic/messages -c 8 -n 1000000 -m 32 -mr 32