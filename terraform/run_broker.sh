git clone https://github.com/polarstreams/polar.git
cd polar && git checkout main && go build .

export POLAR_SHUTDOWN_DELAY_SECS=0
export POLAR_HOME=/data/polar-data
export POLAR_BROKER_NAMES=10.0.0.100,10.0.0.101,10.0.0.102
rm -Rf /data/polar-data && POLAR_ORDINAL=X ./polar
