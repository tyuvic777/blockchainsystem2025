Blockchain Patient Care System

Welcome to the Blockchain Patient Care System, a decentralized healthcare platform built on Hyperledger Fabric, IPFS, and Ethereum. This system provides secure, scalable, and compliant management of patient data, appointments, payments, and telemedicine for admins, doctors, and patients, with role-specific, user-friendly messages.

Overview
The Blockchain Patient Care System integrates:
- Frontend: A Flask-based web interface with AR support, blue/white/green/gray design, and accessibility (ARIA, WCAG 2.1 AA).
- Backend: A Flask-based API with PostgreSQL sharding, caching, and HIPAA/GDPR compliance.
- Blockchain: Hyperledger Fabric for patient care, identity management, and payments, with IPFS for off-chain storage and Ethereum for token rewards.

Role-Specific Messages
- Admin: "Thank you, Admin! Your action on [feature] has been completed successfully."
- Doctor: "Great job, Doctor! Your update to [feature] was successful."
- Patient: "Thank you, [Patient Name]! Your [feature] has been updated successfully."

Features
- Secure patient records and appointments on the blockchain.
- Real-time notifications and telemedicine via WebSocket and AR.
- Token-based reward system for patients and doctors on Ethereum.
- Wearable integration for health monitoring.
- NLP chatbot for patient support.

Prerequisites
- Docker and Docker Compose for blockchain and services.
- Go (version 1.16+) for chaincode development.
- Python (version 3.9+) with Flask, SQLAlchemy, and dependencies.
- Node.js (optional) for frontend dependencies.
