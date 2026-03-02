# Dior - High Performance Data Migration & Press Tool

Dior is a high-performance middleware pressure testing and data migration tool written in Go, supporting Kafka and NSQ.

Its design is inspired by Flume, with core concepts including:
*   **Source**: Responsible for reading data.
*   **Channel**: Buffers data between Source and Sink.
*   **Sink**: Responsible for writing data.

## Features

### Supported Source Types

- 📨 **KafkaSource**: Consume data from Kafka.
- 📬 **NSQSource**: Consume data from NSQ.
- ⚡ **PressSource**: Pressure testing source. Reads a specified file (line by line) and emits data to Sink at a specified rate.

### Supported Sink Types

- 📨 **KafkaSink**: Write data to Kafka.
- 📬 **NSQSink**: Write data to NSQ.
- 💾 **FileSink**: Write data to local file.
- 🗑️ **NilSink**: Empty Sink, used for testing Source performance or discarding data.

## Project Structure

```
dior/
├── cmd/
│   ├── dior/                # Main application entry point
│   │   └── main.go
│   ├── kafka-consumer/      # Kafka consumer utility
│   │   └── main.go
│   └── some/                # Other utility
│       └── main.go
├── component/               # Core component interfaces (Controller, Source, Sink)
│   ├── async.go
│   ├── component.go
│   └── controller.go
├── internal/                # Internal packages
│   ├── cache/               # Data caching logic
│   ├── kafka/               # Kafka related utilities
│   ├── lg/                  # Logging package
│   ├── sink/                # Sink implementations (File, Kafka, NSQ, Nil)
│   ├── source/              # Source implementations (Kafka, NSQ, Press)
│   └── version/             # Version information
├── option/                  # Configuration options and validation
│   ├── option.go
│   └── validate.go
├── Makefile                 # Build script
├── go.mod
└── README.md
```

## Quick Start

### Prerequisites
*   Go 1.20+
*   GNU Make 3.81+

### Build

```bash
make
```

After compilation, binary files will be generated in the `build/` directory.

## Usage Examples

Dior is configured via command-line arguments. Here are some common usage scenarios:

### 1. Press Test Kafka (Press -> Kafka)

Read local file `source.txt` and send to Kafka at a rate of 10 messages per second.

```bash
./build/dior \
  --src press \
  --src-file source.txt \
  --src-speed 10 \
  --dst kafka \
  --dst-bootstrap-servers 127.0.0.1:9092 \
  --dst-topic topic_to
```

### 2. Press Test NSQ (Press -> NSQ)

```bash
./build/dior \
  --src press \
  --src-file source.txt \
  --src-speed 10 \
  --dst nsq \
  --dst-nsqd-tcp-addresses 127.0.0.1:4150 \
  --dst-topic topic_to
```

### 3. Kafka to Kafka Migration

```bash
./build/dior \
  --src kafka \
  --src-bootstrap-servers 127.0.0.1:9092 \
  --src-topic topic_from \
  --src-group benson \
  --dst kafka \
  --dst-bootstrap-servers 127.0.0.1:9092 \
  --dst-topic topic_to
```

### 4. NSQ to NSQ Migration

```bash
./build/dior \
  --src nsq \
  --src-lookupd-tcp-addresses 127.0.0.1:4161 \
  --src-topic topic_from \
  --src-channel benson \
  --dst nsq \
  --dst-nsqd-tcp-addresses 127.0.0.1:4150 \
  --dst-topic topic_to
```

### 5. Kafka to NSQ Migration

```bash
./build/dior \
  --src kafka \
  --src-bootstrap-servers 127.0.0.1:9092 \
  --src-topic topic_from \
  --src-group benson \
  --dst nsq \
  --dst-nsqd-tcp-addresses 127.0.0.1:4150 \
  --dst-topic topic_to
```

### 6. NSQ to Kafka Migration

```bash
./build/dior \
  --src nsq \
  --src-lookupd-tcp-addresses 127.0.0.1:4161 \
  --src-topic topic_from \
  --src-channel benson \
  --dst kafka \
  --dst-bootstrap-servers 127.0.0.1:9092 \
  --dst-topic topic_to
```

### 7. Kafka to File Export

```bash
./build/dior \
  --src kafka \
  --src-bootstrap-servers 127.0.0.1:9092 \
  --src-topic topic_from \
  --src-group benson \
  --dst file \
  --dst-file sink.txt
```

### 8. NSQ to File Export

```bash
./build/dior \
  --src nsq \
  --src-lookupd-tcp-addresses 127.0.0.1:4161 \
  --src-topic topic_from \
  --src-channel benson \
  --dst file \
  --dst-file sink.txt
```

## Development

### Running Tests

```bash
go test ./...
```

### Clean Build

```bash
make clean
```

## Architecture

The application follows a modular architecture:

- **cmd/**: Application entry points.
- **component/**: Defines core interfaces and the Controller which manages the lifecycle of Source and Sink components.
- **internal/**: Contains specific implementations of Sources and Sinks, as well as internal utilities like logging and caching.
- **option/**: Handles configuration parsing and validation.

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
