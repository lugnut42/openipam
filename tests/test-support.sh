#!/bin/bash

# Print test result with colored output
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

# Function to clean up the config directory
cleanup_config_dir() {
    local config_dir="$IPAM_CONFIG_PATH"
    echo "Cleaning up config directory: $config_dir"
    # Remove the entire config directory and recreate it
    rm -rf "$config_dir"
    mkdir -p "$config_dir"
    mkdir -p "$config_dir/blocks"  # Create blocks subdirectory
    print_result $? false
}

# Function to clean up all blocks
cleanup_blocks() {
    echo "Cleaning up existing blocks..."
    # Get all block files in the blocks directory
    for blockfile in "$IPAM_CONFIG_PATH"/blocks/*.yaml; do
        if [ -f "$blockfile" ]; then
            # Get the block name from the filename (remove path and .yaml extension)
            blockname=$(basename "$blockfile" .yaml)
            # Get all CIDR blocks from the file and delete them
            while read -r cidr; do
                if [ ! -z "$cidr" ]; then
                    ./ipam block delete "$cidr" --config "$IPAM_CONFIG_PATH" --force >/dev/null 2>&1
                fi
            done < <(./ipam block list --config "$IPAM_CONFIG_PATH" --file "$blockname" 2>/dev/null | tail -n +2 | awk '{print $1}')
        fi
    done
    print_result $? false
}

# Initialize test environment
init_test_env() {
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

    # Perform initial cleanup
    cleanup_config_dir

    # Initialize the configuration with default block
    echo "Initializing configuration..."
    ./ipam config init default
    print_result $? false

    # Clean up any existing blocks
    cleanup_blocks
}

# Function to perform final cleanup
final_cleanup() {
    echo "Performing final cleanup..."
    cleanup_blocks
    cleanup_config_dir
}