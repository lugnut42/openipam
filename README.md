# OpenIPAM

OpenIPAM is a command-line tool for managing IP address blocks and subnets. It provides a simple, intuitive interface for network administrators to manage their IP address space effectively. The tool features a pattern-based allocation system that simplifies subnet provisioning by allowing administrators to define reusable templates with pre-configured settings for different environments and use cases.

## Table of Contents

- [OpenIPAM](#openipam)
  - [Table of Contents](#table-of-contents)
  - [Getting Started](#getting-started)
  - [Installation](#installation)
    - [Prerequisites](#prerequisites)
    - [From Source](#from-source)
  - [Command Reference](#command-reference)
    - [Configuration Management](#configuration-management)
    - [Block Management](#block-management)
    - [Subnet Management](#subnet-management)
    - [Pattern Management](#pattern-management)
  - [Configuration](#configuration)
  - [Contributing](#contributing)
  - [License](#license)

## Getting Started

1. **Build OpenIPAM**:
   ```bash
   git clone https://github.com/lugnut42/openipam.git
   cd openipam
   make build
   ```

2. **Initialize Configuration**:
   To start off we need to initialize ipam.  This will create a configuration file used by ipam to store the location of block files as well as any patterns we want to define.
   ```bash
    # Initialize with a named block file
    export IPAM_CONFIG_PATH=$HOME/.openipam
    ./ipam config init prod    # Creates blocks/prod.yaml as the initial block file in the .openipam directory
    ```

3. **Add additional block files**:
   After completing initialization, you can add additional block files for ipam to manage.  These block files can be used to represent any construct that works for your and your organization, environmets, regions, data classifications, etc.
   ```
    # Add additional block files to represent environments.  
    ipam config add-block test    # Adds blocks/test.yaml
    ipam config add-block dev     # Adds blocks/dev.yaml
  ```

4. **Assign IP address blocks.**:
  Next we assign IP address blocks to be managed to each of the block files
  ```
    # Use specific block files in commands
    ipam block create --cidr 10.0.0.0/16 --description "Production Network" -f prod
    ipam block create --cidr 172.16.0.0/16 --description "Test Network" -f test
    ipam block create --cidr 192.168.0.0/16 --description "Development Network" -f dev  
   ```

5. **Create a subnet**:
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