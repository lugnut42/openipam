# OpenIPAM

OpenIPAM is a command-line tool for managing IP address blocks and subnets. It provides a simple, intuitive interface for network administrators to manage their IP address space effectively. The tool features a pattern-based allocation system that simplifies subnet provisioning by allowing administrators to define reusable templates with pre-configured settings for different environments and use cases (e.g., development clusters, production VPCs, or service meshes).

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Command Reference](#command-reference)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Quick Start

1. **Install OpenIPAM**:
   ```bash
   go install /openipam@latest
   ```

2. **Initialize configuration**:
   ```bash
   export IPAM_CONFIG_PATH=$HOME/.openipam/ipam-config.yaml
   ipam config init --config $IPAM_CONFIG_PATH
   ```

3. **Create your first IP block**:
   ```bash
   ipam block add --cidr 10.0.0.0/16 --description "Main Datacenter" --file default
   ```

4. **Create a subnet within the block**:
   ```bash
   ipam subnet create --block 10.0.0.0/16 --cidr 10.0.1.0/24 --name "app-tier" --region us-east1 --config $IPAM_CONFIG_PATH
   ```

5. **Create a pattern**:
   ```bash
   ipam pattern create --name dev-gke-uswest --cidr-size 26 --environment dev --region us-west1 --block 10.0.0.0/16 --file default
   ```

6. **Create a subnet using a pattern**:
   ```bash
   ipam subnet create-from-pattern --pattern dev-gke-uswest --file default
   ```

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### From Source

```bash
git clone https:///openipam.git
cd openipam
go build
```

### Using Go Install

```bash
go install /openipam@latest
```

### Enable Shell Completion

For bash:
```bash
ipam completion bash > /etc/bash_completion.d/ipam
```

For zsh:
```bash
ipam completion zsh > "${fpath[1]}/_ipam"
```

## Command Reference

### Core Commands

```bash
ipam                    # Root command
ipam help               # Display help information
ipam completion         # Generate shell completion scripts
```

### Block Management

Manage IP address blocks:

```bash
ipam block add          # Add a new IP block
ipam block list         # List all IP blocks
ipam block show         # Show details of a specific block
ipam block delete       # Delete an IP block
```

### Subnet Management

Manage subnets within IP blocks:

```bash
ipam subnet create      # Create a new subnet
ipam subnet list        # List all subnets
ipam subnet show        # Show details of a specific subnet
ipam subnet delete      # Delete a subnet
ipam subnet create-from-pattern # Create a subnet using a pattern
```

### Pattern Management

Define and manage subnet allocation patterns:

```bash
ipam pattern create     # Create a new pattern
ipam pattern list       # List available patterns
ipam pattern show       # Show pattern details
ipam pattern delete     # Delete a pattern
```

### Configuration

Manage tool configuration:

```bash
ipam config init        # Create initial configuration file
ipam config show        # Display current configuration
```

#### Environment Variables

You must specify the path to the configuration file using the environment variable:

`IPAM_CONFIG_PATH`: Path to the ipam-config.yaml file.

**Example usage:**
```bash
export IPAM_CONFIG_PATH=/path/to/ipam-config.yaml
ipam config show
```
If the `IPAM_CONFIG_PATH` environment variable is not set, the ipam command will throw an error with instructions on how to set it.

## Usage Examples

### Managing IP Blocks

Create a new IP block:
```bash
ipam block add --cidr 10.0.0.0/16 --description "Main Datacenter" --file default
```

List all blocks:
```bash
ipam block list --config $IPAM_CONFIG_PATH
```

Show details of a specific block:
```bash
ipam block show --cidr 10.0.0.0/16 --config $IPAM_CONFIG_PATH
```

Delete an IP block:
```bash
ipam block delete --cidr 10.0.0.0/16 --config $IPAM_CONFIG_PATH --force
```

### Managing Subnets

Create a subnet within a block:
```bash
ipam subnet create --block 10.0.0.0/16 --cidr 10.0.1.0/24 --name "app-tier" --region us-east1 --config $IPAM_CONFIG_PATH
```

List all subnets within a block:
```bash
ipam subnet list --block 10.0.0.0/16 --config $IPAM_CONFIG_PATH
```

Show details of a specific subnet:
```bash
ipam subnet show --cidr 10.0.1.0/24 --config $IPAM_CONFIG_PATH
```

Delete a subnet:
```bash
ipam subnet delete --cidr 10.0.1.0/24 --config $IPAM_CONFIG_PATH --force
```

### Using Patterns

Create a new pattern:
```bash
ipam pattern create --name dev-gke-uswest --cidr-size 26 --environment dev --region us-west1 --block 10.0.0.0/16 --file default
```

List available patterns:
```bash
ipam pattern list --file default
```

Show details of a specific pattern:
```bash
ipam pattern show --name dev-gke-uswest --file default
```

Delete a pattern:
```bash
ipam pattern delete --name dev-gke-uswest --file default
```

Create a subnet using a pattern:
```bash
ipam subnet create-from-pattern --pattern dev-gke-uswest --file default
```

## Configuration

OpenIPAM uses a YAML configuration file located at `$HOME/.openipam/ipam-config.yaml`. Initialize it with:

```bash
ipam config init --config $IPAM_CONFIG_PATH
```

Example configuration:
```yaml
block_files:
  default: /path/to/ip-blocks.yaml
patterns:
  default:
    dev-gke-uswest:
      cidr_size: 26
      environment: dev
      region: us-west1
      block: 10.0.0.0/8
```

## Project Structure

```
openipam/
├── cmd/                 # Command implementations
│   ├── root.go         # Root command and global flags
│   ├── block_cmd.go    # Block management commands
│   ├── subnet_cmd.go   # Subnet management commands
│   └── pattern_cmd.go  # Pattern management commands
│
├── internal/           # Internal packages
│   ├── ipam/           # Core IPAM logic
│   │   ├── ipam_block_add.go     # Block addition logic
│   │   ├── ipam_block_delete.go  # Block deletion logic
│   │   ├── ipam_block_list.go    # Block listing logic
│   │   ├── ipam_block_show.go    # Block showing logic
│   │   ├── ipam_subnet_create.go # Subnet creation logic
│   │   ├── ipam_subnet_delete.go # Subnet deletion logic
│   │   ├── ipam_subnet_list.go   # Subnet listing logic
│   │   ├── ipam_subnet_show.go   # Subnet showing logic
│   │   ├── ipam_pattern.go       # Pattern management logic
│   │   └── util.go               # Helper functions
│   │
│   └── config/        # Configuration management
│       └── config.go    # Config handling
│
└── main.go           # Application entry point
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https:///openipam.git
   ```
3. Create a new branch:
   ```bash
   git checkout -b feature/amazing-feature
   ```
4. Make your changes
5. Run tests:
   ```bash
   go test ./...
   ```
6. Push to your fork and submit a pull request

### Guidelines

- Write tests for new features
- Update documentation as needed
- Follow Go style guidelines
- Write clear commit messages

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [go-cidr](https://github.com/cloudflare/go-cidr) - CIDR manipulation utilities
- Contributors and maintainers

## Version History

- 0.1.0
  - Initial release
  - Basic block and subnet management
  - Pattern-based subnet allocation

## Roadmap

Future enhancements planned for OpenIPAM:

### Version 0.2.0
- Support for alternative storage backends (SQLite, BadgerDB)
- Import functionality from existing spreadsheets
- Subnet utilization reporting

### Version 0.3.0
- REST API for programmatic access
- Web interface for visualization
- Enhanced pattern system with inheritance
- Integration with cloud providers' IPAM systems

### Version 0.4.0
- Multi-user support with role-based access control
- Audit logging
- Subnet request workflow system
- Advanced conflict detection

To request a feature or track progress, please check our [GitHub Issues](https:///openipam/issues).

## Support

Please [open an issue](https:///openipam/issues/new) for support.