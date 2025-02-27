# OpenIPAM

OpenIPAM is a command-line tool for managing IP address blocks and subnets. It provides a simple, intuitive interface for network administrators to manage their IP address space effectively. 

Key features include:
- **Multi-environment management** with separate block files
- **Robust CIDR overlap detection** to prevent IP allocation conflicts
- **Pattern-based subnet allocation** for consistent network provisioning
- **Comprehensive testing** with unit and functional tests
- **Debug capabilities** for troubleshooting

The tool is designed to simplify network administration by allowing administrators to define reusable templates with pre-configured settings for different environments and use cases.

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
    - [Block Files](#block-files)
    - [Patterns](#patterns)
  - [Future enhancements](#future-enhancements)
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
   First, initialize ipam to create a configuration file for storing block file locations and patterns:
   ```bash
   # Set the configuration directory
   export IPAM_CONFIG_PATH=$HOME/.openipam
   
   # Initialize with a named block file (creates blocks/default.yaml as the initial block file)
   ./ipam config init default
   
   # For more verbose output, add the --debug flag
   ./ipam config init default --debug
   ```

3. **Add additional block files**:
   You can add multiple block files to manage different environments or classifications:
   ```bash
   # Add block files for different environments
   ./ipam config add-block prod     # Adds blocks/prod.yaml
   ./ipam config add-block test     # Adds blocks/test.yaml
   ./ipam config add-block dev      # Adds blocks/dev.yaml
   ```

4. **Assign IP address blocks**:
   Define IP blocks for each environment using the --file (-f) flag:
   ```bash
   # Create blocks for different environments
   ./ipam block create --cidr 10.0.0.0/16 --description "Production Network" --file prod
   ./ipam block create --cidr 172.16.0.0/16 --description "Test Network" --file test
   ./ipam block create --cidr 192.168.0.0/16 --description "Development Network" --file dev
   ```

5. **List and inspect blocks**:
   ```bash
   # List all blocks across all files
   ./ipam block list
   
   # List blocks in a specific file
   ./ipam block list --file prod
   
   # Show details of a specific block (now includes utilization statistics)
   ./ipam block show 10.0.0.0/16 --file prod
   
   # See available CIDR ranges within a block
   ./ipam block available 10.0.0.0/16 --file prod
   ```

6. **Create a subnet**:
   ```bash
   # Create a subnet within a block
   ./ipam subnet create --block 10.0.0.0/16 --cidr 10.0.1.0/24 --name "app-tier" --region us-east1
   
   # List subnets
   ./ipam subnet list
   
   # Show subnet details
   ./ipam subnet show --cidr 10.0.1.0/24
   ```

7. **Create and use patterns**:
   ```bash
   # Define a pattern for application subnets
   ./ipam pattern create --name app-subnet --cidr-size 24 --environment prod \
      --region us-east1 --block 10.0.0.0/16
      
   # Create a subnet using the pattern
   ./ipam subnet create-from-pattern --pattern app-subnet
   ```
   
8. **Validate configuration**:
   ```bash
   # Validate block file integrity
   ./validate-blocks prod
   
   # Validate all block files
   ./validate-blocks --all
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
ipam config init [<name>] [--config <path>]
# Creates initial configuration and block file. If name is provided, creates blocks/<name>.yaml
# Default configuration path is $HOME/.openipam/config.yaml

# Add a new block file
ipam config add-block <name> [--config <path>]
# Creates a new block file at blocks/<name>.yaml and adds it to the configuration
# Example: ipam config add-block prod  # Creates blocks/prod.yaml
```

### Block Management

```bash
# Create a new block
ipam block create --cidr <CIDR> [--description <desc>] [--file <key>]

# List all blocks
ipam block list [--file <key>]

# Show block details (includes utilization statistics)
ipam block show <CIDR> [--file <key>]

# Delete a block
ipam block delete <CIDR> [--force] [--file <key>]

# List available CIDR ranges
ipam block available <CIDR> [--file <key>]
```

### Subnet Management

```bash
# Create a subnet
ipam subnet create --block <CIDR> --cidr <CIDR> --name <n> --region <region>

# Create subnet from pattern
ipam subnet create-from-pattern --pattern <n> [--file <key>]

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
ipam pattern create --name <n> --cidr-size <size> --environment <env> --region <region> --block <CIDR> [--file <key>]

# List patterns
ipam pattern list [--file <key>]

# Show pattern details
ipam pattern show --name <n> [--file <key>]

# Delete pattern
ipam pattern delete --name <n> [--file <key>]
```

## Configuration

OpenIPAM uses a YAML-based configuration system with two main components:
1. Configuration file - stores global settings and references to block files
2. Block files - store actual IP blocks, subnets, and patterns for different environments

The configuration file can be specified through:
- Environment variable: `IPAM_CONFIG_PATH` (directory where ipam-config.yaml is stored)
- Command-line flag: `--config <path>` (full path to the config file)
- Default location: `$HOME/.openipam/ipam-config.yaml`

You can also enable debug logging with the `--debug` flag for more detailed output.

### Block Files

Block files store IP blocks, subnets, and patterns. Each block file can represent a different environment, region, or organizational unit. The configuration file references these block files using keys (e.g., 'prod', 'dev', 'test').

Complete configuration example:
```yaml
# Configuration file (config.yaml)
block_files:
  prod: /home/user/.openipam/blocks/prod.yaml
  dev: /home/user/.openipam/blocks/dev.yaml
  test: /home/user/.openipam/blocks/test.yaml
default_block_file: prod  # Default block file to use when --file is not specified

# Example block file (blocks/prod.yaml)
blocks:
  "10.0.0.0/16":
    description: "Production Network"
    subnets:
      "10.0.1.0/24":
        name: "app-tier"
        region: "us-east1"
        description: "Application Tier Subnet"
      "10.0.2.0/24":
        name: "db-tier"
        region: "us-east1"
        description: "Database Tier Subnet"

# Pattern definitions are stored in the config file
patterns:
  web-tier:
    cidr_size: 24
    environment: prod
    region: us-east1
    block: "10.0.0.0/16"
    description: "Web Tier Pattern"
    tags:
      role: web
      environment: production
  app-tier:
    cidr_size: 24
    environment: prod
    region: us-east1
    block: "10.0.0.0/16"
    description: "Application Tier Pattern"
    tags:
      role: application
      environment: production
```

### Patterns

Patterns are templates for subnet creation. They define common settings that can be reused when creating new subnets. Each pattern includes:
- `cidr_size`: The size of the subnet to create (e.g., 24 for a /24 network)
- `environment`: Target environment (e.g., prod, dev, test)
- `region`: Target region for the subnet
- `block`: Parent IP block to allocate from
- `description`: Optional description of the pattern's purpose
- `tags`: Optional key-value pairs for additional metadata

## Features and Capabilities

### Robust CIDR Overlap Detection
OpenIPAM includes intelligent CIDR overlap detection to prevent conflicts:
- Detects partial overlaps between network ranges
- Identifies subnet containment scenarios
- Works across different block files
- Prevents invalid allocations

### Multi-Block File Support
- Manage multiple environments with separate block files
- Reference block files with simple keys
- Apply operations to specific block files with the --file flag

### Pattern-Based Subnet Creation
- Define reusable patterns for common subnet configurations
- Create subnets quickly with predefined parameters
- Maintain consistency across environments

### Subnet Utilization Reporting
- Built-in utilization statistics for blocks and subnets
- Accurate IP address accounting including network/broadcast addresses
- Displays both absolute numbers and percentage utilization
- Helps identify underutilized network segments

### Comprehensive Testing
- Unit tests for all major components
- Functional shell-based tests for CLI operations
- Test coverage for normal and error cases

## Configuration Validation

OpenIPAM includes a standalone validation tool `validate-blocks` that performs comprehensive integrity checks on your block configuration files:

```bash
# Validate the default block file
./validate-blocks

# Validate a specific block file
./validate-blocks prod

# Validate all block files
./validate-blocks --all
```

The validation tool checks for:
- YAML structure and syntax errors
- CIDR format validity
- Subnet containment within parent blocks
- Network overlap detection
- Duplicate resource names and CIDRs
- Required field presence
- Cross-reference integrity

This helps catch configuration errors early and ensures a consistent network design.

## Future enhancements
- Increase test coverage to 100%
- Import / Export functionality
- Cloud Bucket Storage integration
- Pipeline improvements
- Advanced subnet allocation strategies

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please ensure you update tests as appropriate.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.