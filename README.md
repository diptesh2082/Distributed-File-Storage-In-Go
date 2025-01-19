# 🔥 FileStore: Distributed File Storage with P2P Networking

Welcome to FileStore, a powerful distributed file storage system built in Go that leverages peer-to-peer networking for resilient and efficient file management. This project combines robust TCP transport, secure encryption, and content-addressable storage to create a reliable distributed storage solution.
## 🌟 Features

- **📂 Distributed Storage**: Store and retrieve files across a network of peer nodes
- **🔀 P2P Architecture**: Leverage peer-to-peer networking for improved reliability and scalability
- **🔒 Built-in Encryption**: AES encryption keeps your files secure during transmission and storage
- **📍 Content Addressing**: Files are stored using content-addressable storage (CAS) for data integrity
- **🤝 TCP Transport**: Robust TCP-based transport layer for reliable peer communication
- **⚡ Concurrent Operations**: Handle multiple peers and file operations simultaneously
- **🔧 Flexible Configuration**: Customize node behavior, storage paths, and network topology

## 🚀 Getting Started

### Prerequisites

- Go 1.16 or later

### Installation

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/yourusername/filestore.git
   cd filestore
   ```

2. **Build the Project**:
   ```bash
   make build
   ```

3. **Run a Node**:
   ```bash
   make run
   ```

### Usage

- **Starting Multiple Nodes**: The system supports running multiple nodes that automatically discover and connect to each other. See `main.go` for example configurations.
- **Running Tests**:
  ```bash
  make test
  ```

### Core Components

- **Store**: Handles local file storage and retrieval
- **Server**: Manages peer connections and file distribution
- **TCP Transport**: Provides reliable peer-to-peer communication
- **Encryption**: Implements secure file storage and transmission

### Extending FileStore

- **Custom Storage Backends**: Implement alternative storage backends (S3, etc.)
- **Additional Protocols**: Add support for different network protocols
- **Enhanced Discovery**: Implement more sophisticated peer discovery mechanisms
- **Contributing**: We welcome contributions! Please see our contributing guidelines.


## 🏗️ Architecture

FileStore uses a distributed architecture with the following key components:

### Storage Layer
- Content-addressable storage (CAS) for efficient file deduplication
- Encrypted file storage using AES-CTR mode
- Pluggable storage backend interface

### Network Layer 
- TCP-based peer-to-peer communication
- Custom handshake protocol for peer authentication
- Automatic peer discovery and connection management

### Server Layer
- Manages peer connections and file distribution
- Handles file encryption/decryption
- Coordinates storage and network operations

### High-Level Architecture Diagram
### License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details. Feel free to use, modify, and distribute this software as per the terms of the MIT License.

