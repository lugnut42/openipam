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
rm -rf "$IPAM_CONFIG_PATH"
rm -f "$(dirname "$IPAM_CONFIG_PATH")/ipam-blocks.yaml"
print_result $? false

# Check if the configuration file is removed
echo "TEST: Checking if configuration file is removed..."
if [ ! -f "$IPAM_CONFIG_PATH" ]; then
  print_result 0 false
else
  print_result 1 falseß
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
./ipam config init default
print_result $? false

# This should fail because the configuration file already exists
echo "TEST: Re-initializing configuration (should fail)..."
./ipam config init default
print_result $? true

# Test listing with no blocks
echo "TEST: Listing blocks when none exist..."
./ipam block list --config $IPAM_CONFIG_PATH
print_result $? false

# Test available command with no blocks
echo "TEST: Running available command with no blocks (should fail)..."
./ipam block available 10.0.0.0/8
print_result $? true

# Add blocks of different sizes
echo "TEST: Adding /8 block..."
./ipam block create --cidr 10.0.0.0/8 --file default --description "Large block"
print_result $? false

echo "TEST: Adding /16 block..."
./ipam block create --cidr 172.16.0.0/16 --file default --description "Medium block"
print_result $? false

echo "TEST: Adding /24 block..."
./ipam block create --cidr 192.168.1.0/24 --file default --description "Small block"
print_result $? false

# Test creating block without optional parameters
echo "TEST: Adding block without description..."
./ipam block create --cidr 192.168.2.0/24 --file default
print_result $? false

# Test showing blocks with different formats (if supported)
echo "TEST: Showing block details in default format..."
./ipam block show 10.0.0.0/8 --file default --config $IPAM_CONFIG_PATH
print_result $? false

# Test available command on empty block
echo "TEST: Checking available space in unused block..."
./ipam block available 192.168.1.0/24
print_result $? false

# Test available command on non-existent block
echo "TEST: Checking available space in non-existent block (should fail)..."
./ipam block available 172.17.0.0/16
print_result $? true

# List all blocks to verify state
echo "TEST: Listing all blocks..."
./ipam block list --config $IPAM_CONFIG_PATH
print_result $? false

# Test block deletion with --force flag
echo "TEST: Deleting first block with --force flag..."
./ipam block delete 192.168.2.0/24 --config $IPAM_CONFIG_PATH --force
print_result $? false

# Test deleting non-existent block
echo "TEST: Deleting non-existent block (should fail)..."
./ipam block delete 192.168.2.0/24 --config $IPAM_CONFIG_PATH --force
print_result $? true

# Delete remaining blocks with --force
echo "TEST: Deleting remaining blocks..."
# Get list of all blocks and delete them
blocks=$(./ipam block list --config $IPAM_CONFIG_PATH | tail -n +2 | awk '{print $1}')
for block in $blocks; do
    echo "Deleting block: $block"
    ./ipam block delete "$block" --config $IPAM_CONFIG_PATH --force
    print_result $? false
done

# Verify all blocks are deleted
echo "TEST: Verifying all blocks are deleted..."
output=$(./ipam block list --config $IPAM_CONFIG_PATH | tail -n +2)
if [ -z "$output" ]; then
    print_result 0 false
else
    echo "Unexpected output from list command:"
    echo "$output"
    print_result 1 false
fi

# Print overall test result
if [ "$all_tests_passed" = true ]; then
  echo -e "\033[32mALL TESTS PASSED\033[0m"
else
  echo -e "\033[31mSOME TESTS FAILED\033[0m"
fi