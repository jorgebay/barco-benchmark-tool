git clone https://github.com/jorgebay/polar-benchmark-tool.git
cd polar-benchmark-tool && go build .

./polar-benchmark-tool -u http://10.0.0.100:9251/v1/topic/a-topic/messages,http://10.0.0.101:9251/v1/topic/a-topic/messages,http://10.0.0.102:9251/v1/topic/a-topic/messages -c 32 -n 1000000 -m 16 -mr 64 -ch 16