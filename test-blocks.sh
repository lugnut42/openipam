#!/bin/bash

# Initialize overall test status
all_tests_passed=true

# Function to print test result
print_result() {
  local exit_code=$1
  local expected_fail=$2

  if [ $exit_code -eq 0 ]; then
    if [ "$expected_fail" = true ]; then
      echo -e "\033[31mFAIL\033[0m"  # Red for unexpected pass
      all_tests_passed=false
    else
      echo -e "\033[32mPASS\033[0m"  # Green for pass
    fi
  else
    if [ "$expected_fail" = true ]; then
      echo -e "\033[32mPASS\033[0m"  # Green for expected fail
    else
      echo -e "\033[31mFAIL\033[0m"  # Red for fail
      all_tests_passed=false
    fi
  fi
  echo "-----------------------------"
}

# Check if IPAM_CONFIG_PATH is set
if [ -z "$IPAM_CONFIG_PATH" ]; then
  echo -e "\033[31mERROR: IPAM_CONFIG_PATH environment variable is not set. Please set it before running the test.\033[0m"
  exit 1
fi

# Clean up and fresh build
make clean
make build

# Print the IPAM_CONFIG_PATH environment variable
echo "IPAM_CONFIG_PATH: $IPAM_CONFIG_PATH"

# Remove existing configuration and block files
echo "TEST: Removing existing configuration and block files..."
rm -f "$IPAM_CONFIG_PATH"
rm -f "$(dirname "$IPAM_CONFIG_PATH")/ipam-blocks.yaml"
print_result $? false

# Check if the configuration file is removed
echo "TEST: Checking if configuration file is removed..."
if [ ! -f "$IPAM_CONFIG_PATH" ]; then
  print_result 0 false
else
  print_result 1 false
fi

# Check if the block file is removed
echo "TEST: Checking if block file is removed..."
if [ ! -f "$(dirname "$IPAM_CONFIG_PATH")/ipam-blocks.yaml" ]; then
  print_result 0 false
else
  print_result 1 false
fi

# Initialize the configuration
echo "TEST: Initializing configuration..."
./ipam config init --config $IPAM_CONFIG_PATH --block-yaml-file $HOME/.openipam/blocks.yaml
print_result $? false

# This should fail because the configuration file already exists
echo "TEST: Re-initializing configuration (should fail)..."
./ipam config init --config $IPAM_CONFIG_PATH --block-yaml-file $HOME/.openipam/blocks.yaml
print_result $? true

# Add a block to the configuration
echo "TEST: Adding block..."
./ipam block create --cidr 10.0.0.0/8 --file default
print_result $? false

# This should fail because the block already exists
echo "TEST: Adding block again (should fail)..."
./ipam block create --cidr 10.0.0.0/8 --file default
print_result $? true

# Attempt to add a block with an invalid CIDR
echo "TEST: Adding block with invalid CIDR (should fail)..."
./ipam block create --cidr 10.0.0.0/33 --file default
print_result $? true

# Attempt to add a block that overlaps with an existing block
echo "TEST: Adding overlapping block (should fail)..."
./ipam block create --cidr 10.0.0.0/16 --file default
print_result $? true

# Attempt to add a block that is a subset of an existing block
echo "TEST: Adding subset block (should fail)..."
./ipam block create --cidr 10.0.0.0/12 --file default
print_result $? true

# Attempt to add a block that is a superset of an existing block
echo "TEST: Adding superset block (should fail)..."
./ipam block create --cidr 10.0.0.0/7 --file default
print_result $? true

# Attempt to add a block that is adjacent to an existing block
echo "TEST: Adding adjacent block..."
./ipam block create --cidr 11.0.0.0/8 --file default
print_result $? false

# List all blocks
echo "TEST: Listing all blocks..."
./ipam block list --config $IPAM_CONFIG_PATH
print_result $? false

# Show details of an existing block
echo "TEST: Showing block details..."
./ipam block show 10.0.0.0/8 --file default --config $IPAM_CONFIG_PATH
print_result $? false

# Attempt to show details of a non-existent block
echo "TEST: Showing non-existent block details (should fail)..."
./ipam block show 192.168.0.0/16 --file default --config $IPAM_CONFIG_PATH
print_result $? true

# Delete an existing block
echo "TEST: Deleting block..."
./ipam block delete 10.0.0.0/8 --config $IPAM_CONFIG_PATH --force
print_result $? false

# Attempt to delete a non-existent block
echo "TEST: Deleting non-existent block (should fail)..."
./ipam block delete 192.168.0.0/16 --config $IPAM_CONFIG_PATH --force
print_result $? true

# Print overall test result
if [ "$all_tests_passed" = true ]; then
  echo -e "\033[32mALL TESTS PASSED\033[0m"
else
  echo -e "\033[31mSOME TESTS FAILED\033[0m"
fi