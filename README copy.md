# Echo - Decentralized Messaging Platform

A blockchain-anchored messaging platform combining the best features of WhatsApp, Telegram, and Signal with decentralized identity and provable message integrity.

## 🚀 Features

- **Decentralized Identity (DID)** - Self-sovereign identity on Cardano blockchain
- **End-to-End Encryption** - Noise Protocol implementation
- **Blockchain Anchoring** - Provable message integrity via Constellation Hypergraph
- **Trust Scoring** - Dynamic 0-100 trust scores based on behavior and verification
- **ECHO Token Rewards** - Incentive system for platform participation
- **Enterprise Integration** - Verified financial institution communication channels

## 📁 Project Structure

```
echoapp/
├── docs/
│   ├── api/
│   │   └── openapi.yaml      # OpenAPI 3.1 specification
│   ├── architecture/         # Architecture decision records
│   └── PRD.md               # Product Requirements Document
├── src/
│   ├── api/                 # API client and endpoint definitions
│   ├── components/          # Reusable UI components
│   ├── screens/             # Screen/page components
│   ├── services/            # Business logic services
│   │   ├── auth/           # Authentication service
│   │   ├── messaging/      # Messaging service
│   │   ├── identity/       # DID and verification
│   │   ├── crypto/         # Encryption utilities
│   │   └── blockchain/     # Blockchain integration
│   ├── hooks/              # Custom React hooks
│   ├── store/              # State management
│   ├── types/              # TypeScript type definitions
│   └── utils/              # Utility functions
├── config/                  # Configuration files
├── scripts/                 # Build and utility scripts
├── assets/                  # Static assets
│   ├── images/
│   └── fonts/
├── package.json
├── tsconfig.json
└── README.md
```

## 🛠 Tech Stack

### Mobile App
- **Framework**: React Native with Swift Native modules
- **State Management**: Zustand or Redux Toolkit
- **Navigation**: React Navigation
- **Styling**: NativeWind (Tailwind for RN)

### Blockchain & Identity
- **DID Infrastructure**: Cardano (Atala PRISM / Veridian)
- **Messaging Network**: Constellation Hypergraph (DAG)
- **Storage**: IPFS / Filecoin
- **Smart Contracts**: Plutus (Cardano) / Constellation Metagraph

### Security
- **E2E Encryption**: Noise Protocol
- **Key Management**: Device Secure Enclave
- **Authentication**: WebAuthn / Passkeys

## 🚦 Getting Started

### Prerequisites
- Node.js 18+
- React Native CLI
- Xcode (for iOS)
- Android Studio (for Android)
- Cardano wallet (for development)

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/echoapp.git
cd echoapp

# Install dependencies
npm install

# iOS setup
cd ios && pod install && cd ..

# Start Metro bundler
npm start

# Run on iOS
npm run ios

# Run on Android
npm run android
```

## 📖 API Documentation

The API specification is available at `docs/api/openapi.yaml`. You can view it using:

- [Swagger Editor](https://editor.swagger.io/)
- [Redoc](https://redocly.github.io/redoc/)
- VS Code OpenAPI extension

### Generate API Client

```bash
# Using OpenAPI Generator
npx @openapitools/openapi-generator-cli generate \
  -i docs/api/openapi.yaml \
  -g typescript-fetch \
  -o src/api/generated
```

## 🔐 Security

- All messages are end-to-end encrypted using the Noise Protocol
- DIDs are anchored to Cardano blockchain
- Message hashes are stored on Constellation Hypergraph for provability
- Biometric authentication via device Secure Enclave

## 📄 License

Proprietary - All rights reserved

## 🤝 Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution guidelines.
