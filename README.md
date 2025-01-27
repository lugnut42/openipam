# OpenIPAM

OpenIPAM is a command-line tool for managing IP address blocks and subnets. It provides a simple, intuitive interface for network administrators to manage their IP address space effectively. The tool features a pattern-based allocation system that simplifies subnet provisioning by allowing administrators to define reusable templates with pre-configured settings for different environments and use cases.

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Command Reference](#command-reference)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)

## Quick Start

1. **Build OpenIPAM**:
   ```bash
   git clone https://github.com/lugnut42/openipam.git
   cd openipam
   make build
   ```

2. **Initialize configuration**:
   ```bash
   # Set the config path environment variable
   export IPAM_CONFIG_PATH=$HOME/.openipam/ipam-config.yaml

   # Initialize configuration
   ipam config init --config $IPAM_CONFIG_PATH --block-yaml-file $HOME/.openipam/blocks.yaml
   ```

3. **Create an IP block**:
   ```bash
   ipam block create --cidr 10.0.0.0/16 --description "Main Datacenter"
   ```

4. **Create a subnet**:
   ```bash
   ipam subnet create --block 10.0.0.0/16 --cidr 10.0.1.0/24 --name "app-tier" --region us-east1
   ```

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### From Source

```bash
git clone https://github.com/lugnut42/openipam.git
cd openipam
go build
```

## Command Reference

### Configuration Management

```bash
# Initialize configuration
ipam config init --config <path> --block-yaml-file <path>
```

### Block Management

```bash
# Create a new block
ipam block create --cidr <CIDR> [--description <desc>] [--file <key>]

# List all blocks
ipam block list

# Show block details
ipam block show <CIDR> [--file <key>]

# Delete a block
ipam block delete <CIDR> [--force]

# List available CIDR ranges
ipam block available <CIDR> [--file <key>]
```

### Subnet Management

```bash
# Create a subnet
ipam subnet create --block <CIDR> --cidr <CIDR> --name <name> --region <region>

# Create subnet from pattern
ipam subnet create-from-pattern --pattern <name> [--file <key>]

# List subnets
ipam subnet list [--block <CIDR>] [--region <region>]

# Show subnet details
ipam subnet show --cidr <CIDR>

# Delete subnet
ipam subnet delete --cidr <CIDR> [--force]
```

### Pattern Management

```bash
# Create pattern
ipam pattern create --name <name> --cidr-size <size> --environment <env> --region <region> --block <CIDR> [--file <key>]

# List patterns
ipam pattern list [--file <key>]

# Show pattern details
ipam pattern show --name <name> [--file <key>]

# Delete pattern
ipam pattern delete --name <name> [--file <key>]
```

## Configuration

OpenIPAM requires a configuration file specified either through:
- Environment variable: `IPAM_CONFIG_PATH`
- Command-line flag: `--config <path>`

Initialize the configuration:
```bash
ipam config init --config /path/to/config.yaml --block-yaml-file /path/to/blocks.yaml
```

Example configuration structure:
```yaml
block_files:
  default: /path/to/blocks.yaml
patterns:
  default:
    pattern-name:
      cidr_size: 24
      environment: dev
      region: us-west1
      block: 10.0.0.0/16
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please ensure you update tests as appropriate.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.