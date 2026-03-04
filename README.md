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
│   │   ├── main.go          # Main program
│   │   └── main_test.go     # Main program tests
│   └── some/                # Other utility
│       └── main.go
├── component/               # Core component interfaces
│   ├── component.go         # Component interface and factory methods
│   ├── controller.go        # Controller for lifecycle management
│   ├── async.go             # Asynchronizer base component
│   └── async_test.go        # Async tests
├── internal/                # Internal packages
│   ├── cache/               # Data caching logic
│   │   ├── cache.go         # Cache implementation
│   │   └── cache_test.go    # Cache tests
│   ├── lg/                  # Logging package
│   │   ├── logger.go        # Logger interface
│   │   ├── level.go         # Log levels
│   │   ├── appender.go      # Log appenders
│   │   └── std.go           # Standard logger
│   ├── sink/                # Sink implementations
│   │   ├── file.go          # FileSink
│   │   ├── file_test.go     # FileSink tests
│   │   ├── kafka.go         # KafkaSink
│   │   ├── nsq.go           # NSQSink
│   │   └── nil.go           # NilSink
│   ├── source/              # Source implementations
│   │   ├── kafka.go         # KafkaSource
│   │   ├── nsq.go           # NSQSource
│   │   ├── press.go         # PressSource
│   │   └── press_test.go    # PressSource tests
│   └── version/             # Version information
│       └── binary.go        # Version constants
├── option/                  # Configuration options and validation
│   ├── option.go            # Options struct and flag definitions
│   ├── env.go               # Environment variable loading
│   ├── validate.go          # Validation logic
│   └── validate_test.go     # Validation tests
├── docs/
│   └── architecture.md      # Architecture documentation
├── Makefile                 # Build script
├── Dockerfile               # Docker build file
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies checksum
└── README.md                # This file
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

## Cross Compilation

Dior supports cross-compilation to different platforms and architectures. The Makefile automatically detects your current system, but you can override it:

| Target Platform | Command |
|-----------------|---------|
| Linux AMD64 | `make GOOS=linux GOARCH=amd64` |
| Linux ARM64 | `make GOOS=linux GOARCH=arm64` |
| Windows AMD64 | `make GOOS=windows GOARCH=amd64` |
| macOS ARM64 | `make GOOS=darwin GOARCH=arm64` |
| macOS AMD64 | `make GOOS=darwin GOARCH=amd64` |

**注意**：为了确保交叉编译正常工作，请确保在命令行中直接设置环境变量，而不是在 Makefile 中修改。Makefile 会自动检测并应用这些环境变量。

The build system automatically adds `.exe` extension for Windows targets.

To see the current build configuration, use:

```bash
make show-config
```

**示例**：
```bash
# 编译 Linux AMD64 版本
make GOOS=linux GOARCH=amd64

# 编译 Windows AMD64 版本
make GOOS=windows GOARCH=amd64

# 编译 macOS ARM64 版本
make GOOS=darwin GOARCH=arm64
```

## Configuration

Dior can be configured via command-line arguments or environment variables.

### Command Line Options

#### General Options

| Flag | Default | Description |
|------|---------|-------------|
| `--version` | false | Show dior version |
| `--log-level` | info | Log verbosity: debug, info, warn, error, fatal |
| `--log-prefix` | [dior] | Log message prefix |
| `--chan-size` | 100 | Size of queue between source and sink (0 = non-blocking) |

#### Source Options

| Flag | Default | Description |
|------|---------|-------------|
| `--src` | - | Source type: nsq, kafka, press |
| `--src-topic` | - | Source topic (for NSQ/Kafka) |
| `--src-channel` | - | Source channel (for NSQ) |
| `--src-group` | - | Consumer group (for Kafka) |
| `--src-bootstrap-servers` | - | Kafka brokers (comma-separated) |
| `--src-lookupd-http-addresses` | - | NSQ Lookupd HTTP addresses (comma-separated) |
| `--src-nsqd-tcp-addresses` | - | NSQD addresses (comma-separated) |
| `--src-speed` | 10 | Messages per second (for press, 0 = unlimited) |
| `--src-file` | - | Data file path (for press) |
| `--src-scanner-buf-size-mb` | 1 | Scanner buffer size in MB (for press) |

#### Sink Options

| Flag | Default | Description |
|------|---------|-------------|
| `--dst` | - | Destination type: nsq, kafka, file, nil |
| `--dst-topic` | - | Destination topic (for NSQ/Kafka) |
| `--dst-bootstrap-servers` | - | Kafka brokers (comma-separated) |
| `--dst-nsqd-tcp-addresses` | - | NSQD TCP addresses (comma-separated) |
| `--dst-lookupd-http-addresses` | - | NSQ Lookupd HTTP addresses (comma-separated) |
| `--dst-file` | - | Output file path (for file sink) |
| `--dst-buf-size-byte` | 4096 | Write buffer size in bytes (for file sink) |

### Environment Variables

All command-line options can also be set via environment variables (use lowercase with hyphens replaced by underscores):

```bash
# Example: Set source configuration via environment variables
export src=kafka
export src-bootstrap-servers=127.0.0.1:9092
export src-topic=my-topic
export src-group=my-group

# Example: Set destination configuration via environment variables
export dst=file
export dst-file=output.txt
```

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

### 3. Press Test with NilSink (Performance Testing)

Test source performance without actual write operations:

```bash
./build/dior \
  --src press \
  --src-file source.txt \
  --src-speed 100 \
  --dst nil
```

### 4. Kafka to Kafka Migration

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

### 5. NSQ to NSQ Migration

```bash
./build/dior \
  --src nsq \
  --src-lookupd-http-addresses 127.0.0.1:4161 \
  --src-topic topic_from \
  --src-channel benson \
  --dst nsq \
  --dst-nsqd-tcp-addresses 127.0.0.1:4150 \
  --dst-topic topic_to
```

### 6. Kafka to NSQ Migration

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

### 7. NSQ to Kafka Migration

```bash
./build/dior \
  --src nsq \
  --src-lookupd-http-addresses 127.0.0.1:4161 \
  --src-topic topic_from \
  --src-channel benson \
  --dst kafka \
  --dst-bootstrap-servers 127.0.0.1:9092 \
  --dst-topic topic_to
```

### 8. Kafka to File Export

```bash
./build/dior \
  --src kafka \
  --src-bootstrap-servers 127.0.0.1:9092 \
  --src-topic topic_from \
  --src-group benson \
  --dst file \
  --dst-file sink.txt
```

### 9. NSQ to File Export

```bash
./build/dior \
  --src nsq \
  --src-lookupd-http-addresses 127.0.0.1:4161 \
  --src-topic topic_from \
  --src-channel benson \
  --dst file \
  --dst-file sink.txt
```

## Development

### Running Tests

```bash
# Run all tests with verbose output and race detection
make test

# Or run directly
go test -v -race -cover ./...
```

### Clean Build

```bash
make clean
```

### Install to System

```bash
make install
# Installs to /usr/local/bin by default
# Use DESTDIR for custom installation:
make install DESTDIR=/custom/path
```

## Architecture

The application follows a modular architecture:

- **cmd/**: Application entry points.
- **component/**: Defines core interfaces and the Controller which manages the lifecycle of Source and Sink components.
- **internal/**: Contains specific implementations of Sources and Sinks, as well as internal utilities like logging and caching.
- **option/**: Handles configuration parsing and validation.

For detailed architecture documentation, see [docs/architecture.md](docs/architecture.md).

### Core Components

#### Controller
The [`Controller`](component/controller.go:54) manages the lifecycle of Source and Sink components:
- Creates and manages data transfer channel
- Coordinates startup and shutdown of components
- Handles system signals for graceful shutdown
- Ensures data integrity and proper resource cleanup

#### Asynchronizer
The [`Asynchronizer`](component/async.go:45) provides asynchronous processing capabilities:
- Manages data channel reading
- Provides error handling and statistics
- Supports graceful shutdown

#### Component Interface
The [`Component`](component/component.go:12) interface defines the contract for all sources and sinks:
- `Init(channel chan []byte)`: Initialize with data channel
- `Start(ctx context.Context)`: Start processing
- `Stop()`: Stop and cleanup

## Key Features

### Graceful Shutdown
Dior implements a 5-phase graceful shutdown process:
1. Stop Source (stop producing data)
2. Wait for Source goroutines to exit
3. Close Channel (signal Sink no more data)
4. Wait for Sink to drain remaining data
5. Stop Sink (release resources)

### Error Handling
- Panic recovery in all goroutines
- Exponential backoff retry for Kafka operations
- Error counting and statistics
- Configurable error handling callbacks

### Concurrency Safety
- Atomic operations for state management
- Read-write mutex for state access
- WaitGroup for goroutine coordination

## Dependencies

- [github.com/IBM/sarama](https://github.com/IBM/sarama) v1.41.0 - Kafka client library
- [github.com/nsqio/go-nsq](https://github.com/nsqio/go-nsq) v1.1.0 - NSQ client library

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
