Blockchain Patient Care System Deployment Guide

This guide provides step-by-step instructions for deploying the Blockchain Patient Care System, tailored to admins, doctors, and patients with role-specific setup and messages.

Prerequisites
- Docker and Docker Compose (version 1.29+).
- Go (version 1.16+) for chaincode development.
- Python (version 3.9+) with Flask, SQLAlchemy, and dependencies.
- Node.js (optional) for frontend dependencies.
- Ethereum Account: Goerli testnet account with private key for token management.

Installation

1. Clone the Repository
```bash
git clone https://github.com/your-username/BlockchainPatientCareSystem.git
cd BlockchainPatientCareSystem

2. Install Dependencies
Python Dependencies:

pip install -r requirements.txt
Go Dependencies:
cd blockchain/chaincode
go mod tidy

3. Environment Variables: Create a .env file in the root directory with:

DATABASE_URL=postgresql://admin:password@localhost:5432/mediNet
REDIS_HOST=localhost
JWT_SECRET=your-secret-key-here
CSRF_TOKEN=default-csrf-token
PYTHON_BLOCKCHAIN_URL=http://peer0.org1.example.com:7051
ETH_URL=http://localhost:8545
IPFS_URL=http://localhost:5001
ENCRYPTION_KEY=your-encryption-key-here
TWILIO_ACCOUNT_SID=your-twilio-sid
TWILIO_AUTH_TOKEN=your-twilio-token
TWILIO_PHONE_NUMBER=your-twilio-number
EMAIL_USER=your-email@gmail.com
EMAIL_PASSWORD=your-email-password
ETH_PRIVATE_KEY=your-ethereum-private-key
ETH_PRIVATE_KEY_PASSWORD=password

4. Deploy the Blockchain Network

cd blockchain


5. Run the deployment script with role-specific messages:

./deploy-fabric.sh


6. Start Backend Services

python ../backend/app.py


7. Start Frontend Services

python ../frontend/app.py
