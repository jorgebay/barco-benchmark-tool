git clone https://github.com/jorgebay/polar-benchmark-tool.git
cd polar-benchmark-tool && git checkout main && go build .

./polar-benchmark-tool -w binary -hosts 10.0.0.100 -c 6 -n 2000000 -m 1024 -ch 1
