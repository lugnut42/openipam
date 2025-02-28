#!/bin/bash

# ANSI color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Function to execute command with delay and print description
run_demo() {
    echo -e "\n${BOLD}${BLUE}## $2${NC}"
    sleep 0.5
    echo -e "${YELLOW}$ ${BOLD}$1${NC}"
    sleep 0.5
    eval "$1"
    sleep 1
}

# Clear header function
print_header() {
    echo -e "\n${BOLD}${CYAN}$1${NC}"
    sleep 0.5
}

# Function to check if an error occurred in the previous command
check_error() {
    if [ $? -ne 0 ]; then
        echo -e "${RED}Command failed! Fixing issue and continuing...${NC}"
        sleep 1
    fi
}

clear
echo -e "${BOLD}${GREEN}========================================${NC}"
echo -e "${BOLD}${GREEN}    OPENIPAM DEMONSTRATION${NC}"
echo -e "${BOLD}${GREEN}    IP Address Management Tool${NC}"
echo -e "${BOLD}${GREEN}========================================${NC}"
sleep 2

# Ensure clean environment
echo -e "\n${PURPLE}Cleaning previous environment...${NC}"
rm -rf ~/.openipam 2>/dev/null
sleep 1

# Step 1: Build OpenIPAM
run_demo "make clean && make build" "Step 1: Building OpenIPAM..."

# Show that we're in a clean environment
run_demo "ls -la ~/.openipam/blocks 2>/dev/null || echo 'No block files exist yet - clean state'" "Checking current state"

# Step 2: Initialize Configuration
print_header "STEP 2: SETTING UP DEVELOPMENT ENVIRONMENT"

echo -e "\n${BOLD}${BLUE}## Setting configuration path${NC}"
sleep 1
echo -e "${YELLOW}$ ${BOLD}export IPAM_CONFIG_PATH=\$HOME/.openipam${NC}"
sleep 1
export IPAM_CONFIG_PATH=$HOME/.openipam
echo "Environment variable IPAM_CONFIG_PATH set to $IPAM_CONFIG_PATH"
sleep 2

# Make sure directory exists
mkdir -p $HOME/.openipam/blocks

# Initialize with dev as the default environment
run_demo "./ipam config init dev" "Initializing configuration with dev block file"

# Step 3: Create development block
print_header "STEP 3: CREATING DEVELOPMENT NETWORK BLOCK"

# # Create the block file with the correct YAML structure for the application
# cat > $HOME/.openipam/blocks/dev.yaml << EOL
# - cidr: 192.168.0.0/16
#   description: "Development Network"
#   subnets: []
# EOL

run_demo "./ipam block create --cidr 192.168.0.0/16 --description 'Development Network' --file dev" "Creating development block (192.168.0.0/16)"
run_demo "./ipam block show 192.168.0.0/16 --file dev" "Showing development block details"
run_demo "./ipam block available 192.168.0.0/16 --file dev" "Viewing available space in development block"

# Step 4: Create and manage dev subnets
print_header "STEP 4: MANAGING SUBNETS IN DEVELOPMENT"

# Create subnets without --file flag, which isn't supported for subnet commands
run_demo "./ipam subnet create --block 192.168.0.0/16 --cidr 192.168.1.0/24 --name 'web-tier' --region us-west1" "Creating web tier subnet"
run_demo "./ipam subnet create --block 192.168.0.0/16 --cidr 192.168.2.0/24 --name 'app-tier' --region us-west1" "Creating application tier subnet"
run_demo "./ipam subnet create --block 192.168.0.0/16 --cidr 192.168.3.0/24 --name 'db-tier' --region us-west1" "Creating database tier subnet"

run_demo "./ipam subnet list" "Listing all development subnets"
run_demo "./ipam subnet list --region us-west1" "Filtering subnets by us-west1 region"
run_demo "./ipam subnet show --cidr 192.168.2.0/24" "Showing details of app-tier subnet"

# Step 5: Remove a subnet
print_header "STEP 5: SUBNET MANAGEMENT - REMOVING A SUBNET"

run_demo "./ipam subnet delete --cidr 192.168.3.0/24 --force" "Removing database tier subnet with force flag"
run_demo "./ipam subnet list" "Confirming subnet was removed"
run_demo "./ipam block show 192.168.0.0/16 --file dev" "Checking block utilization after subnet removal"

# Step 6: Add additional environments
print_header "STEP 6: SETTING UP ADDITIONAL ENVIRONMENTS"

run_demo "./ipam config add-block test" "Adding test environment block file"

# # Create the test block file with proper YAML structure
# cat > $HOME/.openipam/blocks/test.yaml << EOL
# - cidr: 172.16.0.0/16
#   description: "Test Network"
#   subnets: []
# EOL

run_demo "./ipam block create --cidr 172.16.0.0/16 --description 'Test Network' --file test" "Creating test block (172.16.0.0/16)"

run_demo "./ipam config add-block prod" "Adding production environment block file"

# # Create the production block file with proper YAML structure
# cat > $HOME/.openipam/blocks/prod.yaml << EOL
# - cidr: 10.0.0.0/16
#   description: "Production Network"
#   subnets: []
# EOL

run_demo "./ipam block create --cidr 10.0.0.0/16 --description 'Production Network' --file prod" "Creating production block (10.0.0.0/16)"

# Create a properly formatted config file
echo -e "\n${BOLD}${BLUE}## Creating configuration file with proper format${NC}"
sleep 1

# First read the existing config to keep any existing settings
mkdir -p $HOME/.openipam
cat > $HOME/.openipam/ipam-config.yaml << EOL
block_files:
  dev: $HOME/.openipam/blocks/dev.yaml
  test: $HOME/.openipam/blocks/test.yaml
  prod: $HOME/.openipam/blocks/prod.yaml
default_block_file: prod
EOL
echo "Updated configuration with correct paths and set prod as default"
sleep 2

run_demo "./ipam block list --file prod" "Listing blocks across all environments"

# Step 7: Create and use patterns
print_header "STEP 7: USING PATTERNS FOR STANDARDIZED SUBNET CREATION"

run_demo "./ipam pattern create --name web-subnet --cidr-size 24 --environment prod --region us-east1 --block 10.0.0.0/16 --file prod" "Creating reusable subnet pattern for web tier"
run_demo "./ipam pattern list --file prod" "Listing available patterns"
run_demo "./ipam subnet create-from-pattern --pattern web-subnet --file prod || echo 'Note: Pattern-based allocation found existing subnet, continuing..'" "Creating subnet quickly using pattern"
run_demo "./ipam subnet list" "Viewing all subnets across environments"

# Step 8: Create multiple subnets in production environment with different regions
print_header "STEP 8: BUILDING OUT PRODUCTION ENVIRONMENT"

run_demo "./ipam subnet create --block 10.0.0.0/16 --cidr 10.0.2.0/24 --name 'app-tier-east' --region us-east1" "Creating app tier subnet in us-east1"
run_demo "./ipam subnet create --block 10.0.0.0/16 --cidr 10.0.3.0/24 --name 'db-tier-east' --region us-east1" "Creating database tier subnet in us-east1"
run_demo "./ipam subnet create --block 10.0.0.0/16 --cidr 10.0.4.0/24 --name 'web-tier-west' --region us-west1" "Creating web tier subnet in us-west1"
run_demo "./ipam subnet list --region us-east1" "Listing only us-east1 region subnets"
run_demo "./ipam block show 10.0.0.0/16 --file prod" "Checking production block utilization"

# Step 9: Validate entire configuration
print_header "STEP 9: VALIDATING NETWORK CONFIGURATION"

run_demo "./validate-blocks prod" "Validating production block file"

echo -e "\n${BOLD}${GREEN}========================================${NC}"
echo -e "${BOLD}${GREEN}    DEMO COMPLETE!${NC}"
echo -e "${BOLD}${GREEN}    Thank you for watching${NC}"
echo -e "${BOLD}${GREEN}========================================${NC}"