cd /home/rdso/go/src/github.com/camaison/blockchain-security-layer-in-smart-grid/Blockchain_Configuration/chaincode-goose-temp

docker exec cli peer lifecycle chaincode package goose-temp.tar.gz \
  --path /opt/gopath/src/github.com/hyperledger/chaincode-goose-temp \
  --lang golang \
  --label goose-temp_1

docker exec cli peer lifecycle chaincode install goose-temp.tar.gz 

docker exec -e CORE_PEER_LOCALMSPID="Org2MSP" \
  -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp" \
  -e CORE_PEER_ADDRESS="peer0.org2.example.com:9051" \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" \
  cli peer lifecycle chaincode install goose-temp.tar.gz 

PACKAGE_ID=$(docker exec cli peer lifecycle chaincode queryinstalled | grep "goose-temp_1:" | sed -n 's/.*ID: \([^,]*\), Label.*/\1/p')
echo $PACKAGE_ID


docker exec cli peer lifecycle chaincode checkcommitreadiness \
  --channelID mychannel \
  --name goose-temp \
  --version 1 \
  --sequence 1 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
  --output json


docker exec cli peer lifecycle chaincode approveformyorg --channelID mychannel --name goose-temp --version 1 --package-id $PACKAGE_ID --sequence 1 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --waitForEvent


docker exec cli peer lifecycle chaincode checkcommitreadiness \
  --channelID mychannel \
  --name goose-temp \
  --version 1 \
  --sequence 1 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
  --output json


docker exec -e CORE_PEER_LOCALMSPID="Org2MSP" \
  -e CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp" \
  -e CORE_PEER_ADDRESS="peer0.org2.example.com:9051" \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" \
  cli peer lifecycle chaincode approveformyorg --channelID mychannel --name goose-temp --version 1 --package-id $PACKAGE_ID --sequence 1 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --waitForEvent


docker exec cli peer lifecycle chaincode checkcommitreadiness \
  --channelID mychannel \
  --name goose-temp \
  --version 1 \
  --sequence 1 \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
  --output json


docker exec cli peer lifecycle chaincode commit -o orderer.example.com:7050 --channelID mychannel --name goose-temp --version 1 --sequence 1 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt


docker exec cli peer lifecycle chaincode querycommitted \
  --channelID mychannel \
  --name goose-temp \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem


export CHAINCODE_NAME="goose-temp"

docker exec cli peer chaincode invoke \
-o orderer.example.com:7050 \
--tls \
--cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
-C mychannel -n $CHAINCODE_NAME \
--peerAddresses peer0.org1.example.com:7051 \
--peerAddresses peer0.org2.example.com:9051 \
--tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
--tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
-c '{"function":"InitLedger","Args":[]}'
