# P2P Networking for Validators

This document explains how to use the peer-to-peer (P2P) networking capabilities of the validator node.

## Overview

The validator node uses libp2p for peer-to-peer communication, enabling validators to discover each other and exchange information such as:

- Validator status (registration status, last block seen, etc.)
- Proof submissions
- Data synchronization

This helps validators coordinate their activities and maintain a consistent view of the network.

## Configuration

P2P networking can be configured using the `p2p config` command. The configuration is stored in the `data/p2p_config.json` file.

### Available Configuration Options

- **Listen Addresses**: Network addresses on which the validator listens for incoming connections.
- **Bootstrap Peers**: List of known peers to connect to when starting up.

### Example Configuration

```json
{
  "ListenAddresses": [
    "/ip4/0.0.0.0/tcp/9000",
    "/ip4/0.0.0.0/udp/9000/quic-v1"
  ],
  "BootstrapPeers": [
    "/ip4/192.168.1.100/tcp/9000/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
  ],
  "PrivateKeyFile": ""
}
```

## Commands

### Start Validator with P2P

To start the validator with P2P networking enabled:

```bash
dxp-validator p2p start
```

This will start the validator and enable P2P networking. The validator will listen for incoming connections and attempt to connect to any configured bootstrap peers.

### View P2P Status

To view the current P2P network status:

```bash
dxp-validator p2p status
```

This will display information about the current P2P configuration, including listen addresses and bootstrap peers.

### Configure P2P

To configure P2P networking:

```bash
dxp-validator p2p config --listen "/ip4/0.0.0.0/tcp/9000" --bootstrap "/ip4/192.168.1.100/tcp/9000/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
```

## How It Works

### Peer Discovery

Validators discover each other through two mechanisms:

1. **Local Network Discovery**: Using mDNS, validators can automatically discover other validators on the same local network.
2. **Bootstrap Peers**: Validators can connect to known peers specified in the configuration.

### Message Exchange

Validators exchange messages of different types:

- **Status Messages**: Contain information about the validator's status, including registration status and last block seen.
- **Proof Messages**: Sent when a validator submits a proof to the blockchain.
- **Sync Messages**: Used to request synchronization of data between validators.

### Integration with Validator

The P2P functionality is integrated with the validator's existing functionality:

- When the validator processes a new block, it updates its status and broadcasts it to peers.
- When the validator submits a proof, it broadcasts information about the proof to peers.
- Validators can request synchronization of data from peers if they detect they are behind.

## Troubleshooting

### Common Issues

- **Cannot connect to peers**: Ensure that your firewall allows incoming connections on the configured ports.
- **Peers not discovered on local network**: Make sure that mDNS is not blocked by your network configuration.
- **High CPU usage**: If you experience high CPU usage, try reducing the number of bootstrap peers or increasing the status broadcast interval.

### Logs

P2P networking logs are included in the validator's log output. Look for log entries starting with "P2P:" for P2P-specific information.
