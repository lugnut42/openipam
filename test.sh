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
make

# Print the IPAM_CONFIG_PATH environment variable
echo "IPAM_CONFIG_PATH: $IPAM_CONFIG_PATH"

# Remove existing configuration and block files
echo "Removing existing configuration and block files..."
rm -f "$IPAM_CONFIG_PATH"
rm -f "$(dirname "$IPAM_CONFIG_PATH")/ipam-blocks.yaml"
print_result $? false

# Check if the configuration file is removed
echo "Checking if configuration file is removed..."
if [ ! -f "$IPAM_CONFIG_PATH" ]; then
  print_result 0 false
else
  print_result 1 false
fi

# Check if the block file is removed
echo "Checking if block file is removed..."
if [ ! -f "$(dirname "$IPAM_CONFIG_PATH")/ipam-blocks.yaml" ]; then
  print_result 0 false
else
  print_result 1 false
fi

# Initialize the configuration
echo "Initializing configuration..."
./ipam config init --config $IPAM_CONFIG_PATH
print_result $? false

# This should fail because the configuration file already exists
echo "Re-initializing configuration (should fail)..."
./ipam config init --config $IPAM_CONFIG_PATH
print_result $? true

# Add a block to the configuration
echo "Adding block..."
./ipam block add --cidr 10.0.0.0/8 --file default
print_result $? false

# This should fail because the block already exists
echo "Adding block again (should fail)..."
./ipam block add --cidr 10.0.0.0/8 --file default
print_result $? true

# Attempt to add a block with an invalid CIDR
echo "Adding block with invalid CIDR (should fail)..."
./ipam block add --cidr 10.0.0.0/33 --file default
print_result $? true

# Attempt to add a block that overlaps with an existing block
echo "Adding overlapping block (should fail)..."
./ipam block add --cidr 10.0.0.0/16 --file default
print_result $? true

# Attempt to add a block that is a subset of an existing block
echo "Adding subset block (should fail)..."
./ipam block add --cidr 10.0.0.0/12 --file default
print_result $? true

# Attempt to add a block that is a superset of an existing block
echo "Adding superset block (should fail)..."
./ipam block add --cidr 10.0.0.0/7 --file default
print_result $? true

# Attempt to add a block that is adjacent to an existing block
echo "Adding adjacent block..."
./ipam block add --cidr 11.0.0.0/8 --file default
print_result $? false

# List all blocks
echo "Listing all blocks..."
./ipam block list --config $IPAM_CONFIG_PATH
print_result $? false

# Show details of an existing block
echo "Showing block details..."
./ipam block show 10.0.0.0/8 --file default --config $IPAM_CONFIG_PATH
print_result $? false

# Attempt to show details of a non-existent block
echo "Showing non-existent block details (should fail)..."
./ipam block show 192.168.0.0/16 --file default --config $IPAM_CONFIG_PATH
print_result $? true

# Delete an existing block
echo "Deleting block..."
./ipam block delete 10.0.0.0/8 --config $IPAM_CONFIG_PATH --force
print_result $? false

# Attempt to delete a non-existent block
echo "Deleting non-existent block (should fail)..."
./ipam block delete 192.168.0.0/16 --config $IPAM_CONFIG_PATH --force
print_result $? true

# Subnet tests

# Add a block to the configuration for the subnet tests
echo "Adding block..."
./ipam block add --cidr 10.0.0.0/8 --file default
print_result $? false

# Create a subnet with a valid CIDR within an existing block
echo "Creating subnet..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/24 --name test-subnet --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? false

# Attempt to create a subnet with an invalid CIDR
echo "Creating subnet with invalid CIDR (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/33 --name test-subnet --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? true

# Attempt to create a subnet that overlaps with an existing subnet
echo "Creating overlapping subnet (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/24 --name overlapping-subnet --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? true

# Attempt to create a subnet that is a subset of an existing subnet
echo "Creating subset subnet (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.1.0/25 --name subset-subnet --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? true

# Attempt to create a subnet that is a superset of an existing subnet
echo "Creating superset subnet (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.0.0/23 --name superset-subnet --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? true

# Attempt to create a subnet outside the range of the parent block
echo "Creating subnet outside parent block (should fail)..."
./ipam subnet create --block 10.0.0.0/8 --cidr 11.0.0.0/24 --name outside-subnet --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? true

# Attempt to create a subnet that is adjacent to an existing subnet
echo "Creating adjacent subnet..."
./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.2.0/24 --name adjacent-subnet --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? false

# List all subnets within a block
echo "Listing all subnets..."
./ipam subnet list --block 10.0.0.0/8 --config $IPAM_CONFIG_PATH
print_result $? false

# List subnets filtered by region
echo "Listing subnets by region..."
./ipam subnet list --block 10.0.0.0/8 --region us-east1 --config $IPAM_CONFIG_PATH
print_result $? false

# Show details of an existing subnet
echo "Showing subnet details..."
./ipam subnet show --cidr 10.0.1.0/24 --config $IPAM_CONFIG_PATH
print_result $? false

# Attempt to show details of a non-existent subnet
echo "Showing non-existent subnet details (should fail)..."
./ipam subnet show --cidr 192.168.1.0/24 --config $IPAM_CONFIG_PATH
print_result $? true

# Delete an existing subnet
echo "Deleting subnet..."
./ipam subnet delete --cidr 10.0.1.0/24 --config $IPAM_CONFIG_PATH --force
print_result $? false

# Attempt to delete a non-existent subnet
echo "Deleting non-existent subnet (should fail)..."
./ipam subnet delete --cidr 192.168.1.0/24 --config $IPAM_CONFIG_PATH --force
print_result $? true

# Pattern tests

# Create a pattern with valid parameters
echo "Creating pattern..."
./ipam pattern create --name dev-gke-uswest --cidr-size 26 --environment dev --region us-west1 --block 10.0.0.0/8 --file default
print_result $? false

# Attempt to create a pattern with a name that already exists (should fail)
echo "Creating duplicate pattern (should fail)..."
./ipam pattern create --name dev-gke-uswest --cidr-size 26 --environment dev --region us-west1 --block 10.0.0.0/8 --file default
print_result $? true

# Attempt to create a pattern with an invalid CIDR size (should fail)
echo "Creating pattern with invalid CIDR size (should fail)..."
./ipam pattern create --name invalid-cidr-size --cidr-size 33 --environment dev --region us-west1 --block 10.0.0.0/8 --file default
print_result $? true

# Attempt to create a pattern with a non-existent block (should fail)
echo "Creating pattern with non-existent block (should fail)..."
./ipam pattern create --name non-existent-block --cidr-size 26 --environment dev --region us-west1 --block 192.168.0.0/16 --file default
print_result $? true

# List all patterns for a block file
echo "Listing all patterns..."
./ipam pattern list --file default
print_result $? false

# Show details of an existing pattern
echo "Showing pattern details..."
./ipam pattern show --name dev-gke-uswest --file default
print_result $? false

# Attempt to show details of a non-existent pattern (should fail)
echo "Showing non-existent pattern details (should fail)..."
./ipam pattern show --name non-existent-pattern --file default
print_result $? true

# Create a subnet using a valid pattern
echo "Creating subnet using pattern..."
./ipam subnet create-from-pattern --pattern dev-gke-uswest --file default
print_result $? false

# Fill up the block with subnets
echo "Filling up the block with subnets..."
for i in {1..62}; do
  ./ipam subnet create --block 10.0.0.0/8 --cidr 10.0.3.$((i*4))/30 --name subnet-$i --region us-west1 --config $IPAM_CONFIG_PATH
done

# Attempt to create a subnet using a non-existent pattern (should fail)
echo "Creating subnet using non-existent pattern (should fail)..."
./ipam subnet create-from-pattern --pattern non-existent-pattern --file default
print_result $? true

# Attempt to create a subnet when no available CIDR is left in the block (should fail)
echo "Creating subnet with no available CIDR (should fail)..."
./ipam subnet create-from-pattern --pattern dev-gke-uswest --file default
print_result $? true

# Print overall test result
if [ "$all_tests_passed" = true ]; then
  echo -e "\033[32mALL TESTS PASSED\033[0m"
else
  echo -e "\033[31mSOME TESTS FAILED\033[0m"
fi