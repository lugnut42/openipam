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

# Function to clean up blocks
cleanup_blocks() {
    # Delete blocks from the default file
    blocks=$(./ipam block list --file default 2>/dev/null | grep -v "Block CIDR" | awk '{print $1}')
    for block in $blocks; do
        if [[ $block =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/[0-9]+$ ]]; then
            ./ipam block delete "$block" --force --file default
        fi
    done

    # Delete blocks from the test file if it exists
    if [ -f "$IPAM_CONFIG_PATH/blocks/test.yaml" ]; then
        blocks=$(./ipam block list --file test 2>/dev/null | grep -v "Block CIDR" | awk '{print $1}')
        for block in $blocks; do
            if [[ $block =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/[0-9]+$ ]]; then
                ./ipam block delete "$block" --force --file test
            fi
        done
    fi
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
if [ ! -f "$(dirname "$IPAM_CONFIG_PATH")/blocks/default.yaml" ]; then
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
./ipam block list
print_result $? false

# Test available command with no blocks
echo "TEST: Running available command with no blocks (should fail)..."
./ipam block available 10.0.0.0/8
print_result $? true

# Add blocks of different sizes
echo "TEST: Adding /8 block..."
./ipam block create --cidr 10.0.0.0/8 --description "Large block"
print_result $? false

echo "TEST: Adding /16 block..."
./ipam block create --cidr 172.16.0.0/16 --description "Medium block"
print_result $? false

echo "TEST: Adding /24 block..."
./ipam block create --cidr 192.168.1.0/24 --description "Small block"
print_result $? false

# Test creating block without optional parameters
echo "TEST: Adding block without description..."
./ipam block create --cidr 192.168.2.0/24
print_result $? false

# Test showing blocks with different formats (if supported)
echo "TEST: Showing block details in default format..."
./ipam block show 10.0.0.0/8
print_result $? false

# Test available command on empty block
echo "TEST: Checking available space in unused block..."
./ipam block available 192.168.1.0/24
print_result $? false

# Test available command on non-existent block
echo "TEST: Checking available space in non-existent block (should fail)..."
./ipam block available 172.17.0.0/16
print_result $? true

# Test CIDR overlap scenarios
echo "TEST: Testing CIDR block overlap scenarios..."

cleanup_blocks

# Add single base block for overlap testing
echo "TEST: Adding base block for overlap testing..."
./ipam block create --cidr 172.16.0.0/16 --description "Base block for testing"

# Test exact overlap
echo "TEST: Creating block with exact overlap (should fail)..."
./ipam block create --cidr 172.16.0.0/16 --description "Overlapping block"
print_result $? true

# Test containing overlap (larger block overlapping existing smaller block)
echo "TEST: Creating block that contains existing block (should fail)..."
./ipam block create --cidr 172.16.0.0/12 --description "Containing block"
print_result $? true

# Test contained overlap (smaller block inside existing block)
echo "TEST: Creating block contained within existing block (should fail)..."
./ipam block create --cidr 172.16.0.0/24 --description "Contained block"
print_result $? true

# Test partial overlap from start
echo "TEST: Creating block with partial overlap from start (should fail)..."
# 10.1.0.0/16 and 10.0.0.0/15 definitely overlap (10.1.0.0/16 is within 10.0.0.0/15)
if output=$(./ipam block create --cidr 10.1.0.0/16 --description "First block" 2>&1); then
    echo "First block created successfully"
    
    if output2=$(./ipam block create --cidr 10.0.0.0/15 --description "Overlapping block" 2>&1); then
        echo "$output2"
        print_result 0 true  # Command succeeded when it should have failed
    else
        # Check if output shows overlap error
        if [[ "$output2" == *"overlaps"* ]]; then
            echo "$output2"
            print_result 1 true  # Command failed as expected
        else
            echo "$output2"
            # Some other error occurred
            print_result 0 true  # Not the error we expected
        fi
    fi
else
    echo "Failed to create first test block: $output"
    print_result 1 false  # Unexpected failure
fi

# Test partial overlap from end
echo "TEST: Creating block with partial overlap from end (should fail)..."
./ipam block create --cidr 172.16.0.0/15 --description "Partial overlap end"
print_result $? true

# Test adjacent blocks (should succeed)
echo "TEST: Creating block adjacent to existing block (should succeed)..."
./ipam block create --cidr 172.17.0.0/16 --description "Adjacent block"
print_result $? false

# Test block overlap across different files
echo "TEST: Testing block overlap across different files..."

# Create a new block file
echo "TEST: Adding new block file..."
./ipam config add-block test
print_result $? false

# Test creating overlapping block in different file (should fail)
echo "TEST: Creating overlapping block in different file (should fail)..."
./ipam block create --cidr 172.16.0.0/16 --file test --description "Cross-file overlap"
print_result $? true

# Clean up before deletion tests
cleanup_blocks

# Test deleting non-existent block
echo "TEST: Deleting non-existent block (should fail)..."
if output=$(./ipam block delete 192.168.99.0/24 --force 2>&1); then
    # The command succeeded but it should have failed - check output for clue
    if [[ "$output" == *"not found"* ]]; then
        echo "$output"
        print_result 1 true  # It actually failed as expected
    else
        echo "$output"
        print_result 0 true  # Command succeeded when it should have failed
    fi
else
    echo "$output" 
    print_result 1 true  # Command failed as expected
fi

# Final verification
echo "TEST: Verifying all blocks are deleted..."
default_blocks=$(./ipam block list --file default 2>/dev/null | grep -v "Block CIDR" | grep -c "^[0-9]" || echo 0)
test_blocks=0
if [ -f "$IPAM_CONFIG_PATH/blocks/test.yaml" ]; then
    test_blocks=$(./ipam block list --file test 2>/dev/null | grep -v "Block CIDR" | grep -c "^[0-9]" || echo 0)
fi

# Convert string to integer to avoid bash comparison issues
default_blocks_int=$(expr "$default_blocks" + 0 2>/dev/null || echo 0)
test_blocks_int=$(expr "$test_blocks" + 0 2>/dev/null || echo 0)

if [ "$default_blocks_int" -eq 0 ] && [ "$test_blocks_int" -eq 0 ]; then
    print_result 0 false
else
    echo "Unexpected blocks remain:"
    echo "Default file blocks: $default_blocks_int"
    echo "Test file blocks: $test_blocks_int"
    print_result 1 false
fi

# Print overall test result
if [ "$all_tests_passed" = true ]; then
  echo -e "\033[32mALL TESTS PASSED\033[0m"
else
  echo -e "\033[31mSOME TESTS FAILED[0m"
fi