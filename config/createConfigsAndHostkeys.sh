#!/bin/bash
n=200

echo "started creation of ${n} peers"

for ((i = 1; i <= ${n}; ++i)); do
let  "p2pPort = 3000 + ${i} * 2"
let  "apiPort = ${p2pPort} - 1"
let  "neighbor1 =  3000 + (${i}*2 +2 ) % ( ${n} * 2 )"  
let  "neighbor2 = 3000 + (${i}*2 +4 ) % ( ${n} * 2 )" 
let  "neighbor3 =  3000 + (${i}*2 +6 ) % ( ${n} * 2 )"  
echo ${apiPort}
openssl genrsa -out hostkey${i}.pem 4096   

  cat > "config${i}.ini" << EOF 

hostkey = config/hostkey${i}.pem

[dht]
api_address = 127.0.0.1:${apiPort}
p2p_address = 127.0.0.1:${p2pPort}
maxTTL = 86400
preConfPeer1 = 127.0.0.1:${neighbor1}
preConfPeer2 = 127.0.0.1:${neighbor2}
preConfPeer3 = 127.0.0.1:${neighbor3}
k = 5
a = 3
EOF

done
