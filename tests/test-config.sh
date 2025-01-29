#!/bin/bash

source ./tests/test-support.sh

# Initialize overall test status
all_tests_passed=true

# Initialize test environment
init_test_env

# Test adding a new block file
echo "TEST: Adding new block file 'production'..."
./ipam config add-block production
print_result $? false

# Test adding duplicate block file (should fail)
echo "TEST: Adding duplicate block file 'production' (should fail)..."
./ipam config add-block production
print_result $? true

# Test adding block file with special characters (should fail)
echo "TEST: Adding block file with special characters (should fail)..."
./ipam config add-block "test/block"
print_result $? true

# Test adding block file with only valid characters
echo "TEST: Adding block file with valid characters..."
./ipam config add-block "test_block"
print_result $? false

# Test adding block file with hyphens
echo "TEST: Adding block file with hyphens..."
./ipam config add-block "test-block-2"
print_result $? false

# Test adding block file with spaces (should fail)
echo "TEST: Adding block file with spaces (should fail)..."
./ipam config add-block "test block"
print_result $? true

# Test if block files exist in correct locations
echo "TEST: Verifying block file locations..."
if [ -f "$IPAM_CONFIG_PATH/blocks/test_block.yaml" ] && \
   [ -f "$IPAM_CONFIG_PATH/blocks/test-block-2.yaml" ] && \
   [ -f "$IPAM_CONFIG_PATH/blocks/production.yaml" ]; then
    print_result 0 false
else
    print_result 1 false
fi

# Clean up
final_cleanup

# Print overall test result
if [ "$all_tests_passed" = true ]; then
    echo -e "\033[32mALL TESTS PASSED\033[0m"
else
    echo -e "\033[31mSOME TESTS FAILED\033[0m"
fi