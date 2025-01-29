#!/bin/bash

source ./tests/test-support.sh

# Initialize overall test status
all_tests_passed=true

# Initialize test environment
init_test_env

# Add multiple blocks for testing
echo "Adding first block..."
./ipam block create --cidr 10.0.0.0/8 --file default --description "Main test block"
print_result $? false

echo "Adding second block..."
./ipam config add-block test
./ipam block create --cidr 172.16.0.0/12 --file test --description "Secondary test block"
print_result $? false

# Test listing subnets when none exist
echo "Listing subnets with empty block..."
./ipam subnet list --block 10.0.0.0/8
print_result $? false

# Create subnets in different regions
echo "Creating subnet in us-east1..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/24 --name test-subnet-east --region us-east1
print_result $? false

echo "Creating subnet in us-west1..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.2.0/24 --name test-subnet-west --region us-west1
print_result $? false

echo "Creating subnet in eu-west1..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.3.0/24 --name test-subnet-eu --region eu-west1
print_result $? false

# Create a subnet in the second block
echo "Creating subnet in second block..."
./ipam subnet create --block 172.16.0.0/12 --cidr 172.16.1.0/24 --name test-subnet-secondary --region us-east1
print_result $? false

# Test invalid cases
echo "Creating subnet with invalid CIDR (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/33 --name test-subnet --region us-east1
print_result $? true

echo "Creating overlapping subnet (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/24 --name overlapping-subnet --region us-east1
print_result $? true

echo "Creating subset subnet (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/25 --name subset-subnet --region us-east1
print_result $? true

echo "Creating superset subnet (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.0.0/23 --name superset-subnet --region us-east1
print_result $? true

echo "Creating subnet outside parent block (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 11.0.0.0/24 --name outside-subnet --region us-east1
print_result $? true

# Test listing scenarios
echo "Listing all subnets (multiple blocks)..."
./ipam subnet list
print_result $? false

echo "Listing subnets for specific block..."
./ipam subnet list --block 10.0.0.0/8
print_result $? false

echo "Listing subnets by region (us-east1)..."
./ipam subnet list --block 10.0.0.0/8 --region us-east1
print_result $? false

echo "Listing subnets by region (us-west1)..."
./ipam subnet list --block 10.0.0.0/8 --region us-west1
print_result $? false

# Test show command scenarios
echo "Showing subnet details (with all fields populated)..."
./ipam subnet show --cidr 10.0.1.0/24
print_result $? false

echo "Showing subnet from secondary block..."
./ipam subnet show --cidr 172.16.1.0/24
print_result $? false

echo "Showing non-existent subnet details (should fail)..."
./ipam subnet show --cidr 192.168.1.0/24
print_result $? true

# Test delete scenarios
echo "Attempting delete without --force flag..."
./ipam subnet delete --cidr 10.0.3.0/24
print_result $? true

echo "Deleting subnet with --force flag..."
./ipam subnet delete --cidr 10.0.3.0/24 --force
print_result $? false

echo "Deleting subnet from secondary block..."
./ipam subnet delete --cidr 172.16.1.0/24 --force
print_result $? false

echo "Deleting non-existent subnet (should fail)..."
./ipam subnet delete --cidr 192.168.1.0/24 --force
print_result $? true

# Delete remaining subnets
echo "Deleting remaining subnets..."
./ipam subnet delete --cidr 10.0.1.0/24 --force
print_result $? false
./ipam subnet delete --cidr 10.0.2.0/24 --force
print_result $? false

# Verify empty state
echo "Verifying all subnets are deleted..."
# Skip the header row and check that we have no data rows
output=$(./ipam subnet list | tail -n +2 | grep -v "^Block CIDR")
if [ -z "$output" ]; then
    print_result 0 false
else
    echo "Unexpected output from list command:"
    echo "$output"
    print_result 1 false
fi

# Clean up the test blocks
echo "Final cleanup..."
cleanup_blocks
print_result $? false

# Print overall test result
if [ "$all_tests_passed" = true ]; then
  echo -e "\033[32mALL TESTS PASSED\033[0m"
else
  echo -e "\033[31mSOME TESTS FAILED\033[0m"
fi