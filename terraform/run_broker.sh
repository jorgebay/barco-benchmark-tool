git clone https://github.com/barcostreams/barco.git
cd barco && go build .

export BARCO_SHUTDOWN_DELAY_SECS=0
export BARCO_HOME=/data/barco-data
export BARCO_BROKER_NAMES=10.0.0.100,10.0.0.101,10.0.0.102
rm -Rf /data/barco-data && BARCO_ORDINAL=X ./barco
