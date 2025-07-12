# DHCP Security Event Streaming Platform

A robust, real-time event streaming platform built in Go for monitoring, detecting, and responding to DHCP(Dynamic Host Configuration Protoco)-related security threats across enterprise networks. Leveraging Apache Kafka for scalable event transport, this platform ingests DHCP security events, applies advanced detection logic to identify attacks (such as starvation, rogue servers, and spoofing), and generates actionable security alerts. It features modular producers and consumers, a flexible event processing engine, and a built-in simulator for testing and demonstration. Designed for network security teams and researchers, this solution enables proactive defense and visibility into DHCP-based attack vectors.

---

## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Usage](#usage)
- [Configuration](#configuration)
- [Event & Alert Schema](#event--alert-schema)
- [Contributing](#contributing)
- [License](#license)

---

## Overview
This project provides a streaming platform for monitoring DHCP-related security events in real time. It leverages Apache Kafka for event streaming and Go for high-performance event processing. The platform can:
- Collect DHCP security events from network monitoring tools
- Process and analyze events to detect attacks (e.g., starvation, rogue servers, spoofing)
- Generate and publish security alerts
- Simulate DHCP events for testing/demo purposes

## Features
- **Kafka Producer/Consumer:** Reliable event ingestion and processing using [sarama](https://github.com/Shopify/sarama)
- **Event Processing:** Detects DHCP attacks using configurable rules
- **Security Alerts:** Publishes alerts to a dedicated Kafka topic
- **Simulation:** Built-in simulator for generating synthetic DHCP events
- **Graceful Shutdown:** Handles SIGINT/SIGTERM for safe shutdown

## Architecture
```
[Network Monitoring Tools] --(DHCP Events)--> [Kafka Topic: dhcp-security-events]
      |                                                        |
      |                                                [Consumer]
      |                                                        |
      +--<--<--<--<--<--<--<--<--<--<--<--<--<--<--<--<--<--<--<--+
                                                           |
                                                    [Event Processor]
                                                           |
                                                [Kafka Topic: dhcp-security-alerts]
                                                           |
                                                    [Alert Consumers]
```

## Prerequisites
- Go 1.23+
- Apache Kafka (local or remote cluster)
- [sarama](https://github.com/Shopify/sarama) Go client (managed via `go.mod`)

## Setup
1. **Clone the repository:**
   ```sh
   git clone <repo-url>
   cd dhcp-monitoring-app
   ```
2. **Install dependencies:**
   ```sh
   go mod tidy
   ```
3. **Start Kafka:**
   - Ensure Kafka is running and accessible at the address specified in the config (default: `localhost:9092`).
   - Create the following topics (if not auto-created):
     - `dhcp-security-events`
     - `dhcp-security-alerts`

## Usage
### Run the Platform
```sh
go run main.go
```
- The platform will start consuming DHCP events, process them, and publish alerts.
- By default, a simulator will generate synthetic DHCP events for demonstration.

### Stopping
- Press `Ctrl+C` to gracefully shut down the platform.

## Configuration
Configuration is set in `main.go` via the `KafkaConfig` struct:
```go
config := KafkaConfig{
    Brokers:         []string{"localhost:9092"},
    DHCPEventsTopic: "dhcp-security-events",
    AlertsTopic:     "dhcp-security-alerts",
    ConsumerGroup:   "dhcp-security-processors",
}
```
- **Brokers:** List of Kafka broker addresses
- **DHCPEventsTopic:** Topic for incoming DHCP events
- **AlertsTopic:** Topic for security alerts
- **ConsumerGroup:** Kafka consumer group ID

## Event & Alert Schema
### DHCP Security Event
```json
{
  "id": "event_1",
  "timestamp": "2024-06-01T12:00:00Z",
  "event_type": "DHCP_DISCOVER",
  "severity": "LOW",
  "source_ip": "192.168.1.10",
  "dest_ip": "192.168.1.1",
  "client_mac": "aa:bb:cc:dd:ee:01",
  "server_mac": "00:11:22:33:44:55",
  "transaction_id": 12345,
  "requested_ip": "192.168.1.100",
  "offered_ip": "192.168.1.100",
  "lease_time": 3600,
  "network_id": "network_001",
  "description": "DHCP DISCOVER event from monitoring tool",
  "metadata": { "monitoring_tool": "NetworkAnalyzer", "interface": "eth0", "vlan_id": 100 }
}
```

### Security Alert
```json
{
  "id": "alert_1234567890",
  "timestamp": "2024-06-01T12:01:00Z",
  "alert_type": "DHCP Starvation Attack",
  "severity": "HIGH",
  "source": "192.168.1.10",
  "description": "Potential DHCP starvation attack detected",
  "events": [ ... ],
  "metadata": { "rule_name": "DHCP Starvation Attack", "network_id": "network_001", "client_mac": "...", "server_mac": "..." }
}
```

## Contributing
Contributions are welcome! Please open issues or submit pull requests for improvements, bug fixes, or new features.

## License
[MIT](LICENSE)
