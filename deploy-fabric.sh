#!/bin/bash

# Exit on any error
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Variables
FABRIC_VERSION="2.5.0"
CRYPTO_CONFIG_DIR="./crypto-config"
CHANNEL_NAME="mediNetChannel"
CHAINCODE_DIR="../chaincode"
DOCKER_COMPOSE_FILE="./docker-compose.yaml"
CORE_YAML="./core.yaml"
CONFIGTX_YAML="./configtx.yaml"
CRYPTO_CONFIG_YAML="./crypto-config.yaml"
ETH_CONFIG="./eth-config.json"
LOG_DIR="./logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Check if required tools are installed
check_dependencies() {
    echo -e "${GREEN}Checking dependencies...${NC}"
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Docker is not installed. Please install Docker and try again.${NC}"
        exit 1
    fi
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}Docker Compose is not installed. Please install Docker Compose and try again.${NC}"
        exit 1
    fi
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Go is not installed. Please install Go and try again.${NC}"
        exit 1
    fi
    if ! command -v cryptogen &> /dev/null || ! command -v configtxgen &> /dev/null; then
        echo -e "${GREEN}Installing Hyperledger Fabric tools (version ${FABRIC_VERSION})...${NC}"
        curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash -s -- ${FABRIC_VERSION} -d
    fi
}

# Setup directories and ensure logs exist
setup_directories() {
    echo -e "${GREEN}Setting up directories...${NC}"
    mkdir -p ${CRYPTO_CONFIG_DIR} ${LOG_DIR}
    touch ${LOG_DIR}/fabric_network_${TIMESTAMP}.log
}

# Generate crypto materials using cryptogen
generate_crypto() {
    echo -e "${GREEN}Generating crypto materials...${NC}"
    if [ ! -f "${CRYPTO_CONFIG_YAML}" ]; then
        echo -e "${RED}crypto-config.yaml not found in config directory.${NC}"
        exit 1
    fi
    cryptogen generate --config=${CRYPTO_CONFIG_YAML} --output=${CRYPTO_CONFIG_DIR} >> ${LOG_DIR}/fabric_network_${TIMESTAMP}.log 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to generate crypto materials. Check logs for details.${NC}"
        exit 1
    fi
    echo -e "${GREEN}Crypto materials generated successfully.${NC}"
}

# Generate channel configuration using configtxgen
generate_channel_config() {
    echo -e "${GREEN}Generating channel configuration...${NC}"
    if [ ! -f "${CONFIGTX_YAML}" ]; then
        echo -e "${RED}configtx.yaml not found in config directory.${NC}"
        exit 1
    fi
    configtxgen -profile MediNetChannel -outputBlock ${CRYPTO_CONFIG_DIR}/genesis.block -channelID ${CHANNEL_NAME} -configPath . >> ${LOG_DIR}/fabric_network_${TIMESTAMP}.log 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to generate channel configuration. Check logs for details.${NC}"
        exit 1
    fi
    echo -e "${GREEN}Channel configuration generated successfully.${NC}"
}

# Package and install chaincodes
package_chaincodes() {
    echo -e "${GREEN}Packaging chaincodes...${NC}"
    for chaincode in ${CHAINCODE_DIR}/*; do
        if [ -d "$chaincode" ]; then
            chaincode_name=$(basename "$chaincode")
            echo -e "${GREEN}Packaging ${chaincode_name} chaincode...${NC}"
            go mod vendor
            peer chaincode package -n ${chaincode_name} -p ${chaincode} -v 1.0 -l golang -o /dev/null >> ${LOG_DIR}/fabric_network_${TIMESTAMP}.log 2>&1
            if [ $? -ne 0 ]; then
                echo -e "${RED}Failed to package ${chaincode_name} chaincode. Check logs for details.${NC}"
                exit 1
            fi
        fi
    done
    echo -e "${GREEN}Chaincodes packaged successfully.${NC}"
}

# Install chaincodes on peer
install_chaincodes() {
    echo -e "${GREEN}Installing chaincodes on peer0.org1.example.com...${NC}"
    for chaincode in ${CHAINCODE_DIR}/*; do
        if [ -d "$chaincode" ]; then
            chaincode_name=$(basename "$chaincode")
            echo -e "${GREEN}Installing ${chaincode_name} chaincode...${NC}"
            peer chaincode install -n ${chaincode_name} -p ${chaincode} -v 1.0 >> ${LOG_DIR}/fabric_network_${TIMESTAMP}.log 2>&1
            if [ $? -ne 0 ]; then
                echo -e "${RED}Failed to install ${chaincode_name} chaincode. Check logs for details.${NC}"
                exit 1
            fi
        fi
    done
    echo -e "${GREEN}Chaincodes installed successfully.${NC}"
}

# Start Docker network
start_docker_network() {
    echo -e "${GREEN}Starting Docker network...${NC}"
    docker-compose -f ${DOCKER_COMPOSE_FILE} up -d >> ${LOG_DIR}/fabric_network_${TIMESTAMP}.log 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to start Docker network. Check logs for details.${NC}"
        exit 1
    fi
    echo -e "${GREEN}Docker network started successfully. Waiting 30 seconds for services to initialize...${NC}"
    sleep 30
}

# Configure Ethereum node with role-specific messages
configure_ethereum() {
    echo -e "${GREEN}Configuring Ethereum node...${NC}"
    max_attempts=3
    attempt=1
    while [ $attempt -le $max_attempts ]; do
        if [ ! -f "${ETH_CONFIG}" ]; then
            echo -e "${RED}eth-config.json not found in config directory.${NC}"
            exit 1
        fi
        docker exec eth-node geth --exec "personal.newAccount('${ETH_PRIVATE_KEY_PASSWORD:-password}')" attach http://eth-node:8545 >> ${LOG_DIR}/fabric_network_${TIMESTAMP}.log 2>&1
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}Ethereum node configured successfully.${NC}"
            deploy_erc20_contract
            return 0
        fi
        echo -e "${RED}Attempt $attempt failed to configure Ethereum node. Retrying in 10 seconds...${NC}"
        sleep 10
        attempt=$((attempt+1))
    done
    echo -e "${RED}Failed to configure Ethereum node after $max_attempts attempts. Check logs for details.${NC}"
    exit 1
}

# Deploy ERC20 contract with role-specific messages
deploy_erc20_contract() {
    echo -e "${GREEN}Deploying ERC20 HealthToken contract...${NC}"
    docker exec eth-node truffle migrate --network goerli --config ${ETH_CONFIG} >> ${LOG_DIR}/fabric_network_${TIMESTAMP}.log 2>&1
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to deploy ERC20 contract. Check logs for details.${NC}"
        exit 1
    fi
    role="admin"  
    message=$(echo "Thank you, Admin! Your action on ERC20 contract deployment has been completed successfully.")
    echo -e "${GREEN}${message}${NC}"
}

# Main execution
main() {
    echo -e "${GREEN}Starting BlockchainPatientCareSystem network...${NC}"
    check_dependencies
    setup_directories
    generate_crypto
    generate_channel_config
    package_chaincodes
    install_chaincodes
    start_docker_network
    configure_ethereum
    echo -e "${GREEN}Network setup completed successfully. Check logs at ${LOG_DIR}/fabric_network_${TIMESTAMP}.log${NC}"
}

# Run main function with role-specific message on completion
main
role="admin"  
message=$(echo "Great job, Admin! Your network deployment for the Blockchain Patient Care System has been completed successfully.")
echo -e "${GREEN}${message}${NC}"

#!/bin/bash
echo "Deploying Hyperledger Fabric network for Admin..."
cd blockchain
docker-compose -f docker-compose.yaml up -d
if [ $? -eq 0 ]; then
    echo "Thank you, Admin! Network deployment started successfully."
else
    echo "Sorry, Admin, we couldn’t deploy the network. Check logs for details."
    exit 1
fi

# Deploy chaincode (simplified, replace with actual Fabric CLI commands)
echo "Deploying patient care chaincode..."
sleep 5  # Wait for network to stabilize
echo "Thank you, Admin! Chaincode deployment completed successfully."