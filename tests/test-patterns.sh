#!/bin/bash

source ./tests/test-support.sh

# Initialize test tracking
all_tests_passed=true

# Initialize test environment
init_test_env

# Add a block to the configuration
echo "TEST: Adding block..."
./ipam block create --cidr 10.0.0.0/8 --file default --description "Large block"
print_result $? false

# Create a pattern with valid parameters
echo "TEST: Creating pattern..."
./ipam pattern create --name dev-gke-uswest --cidr-size 26 \
    --environment dev --region us-west1 \
    --block 10.0.0.0/8 --file default
print_result $? false

# Attempt to create a pattern with a name that already exists (should fail)
echo "TEST: Creating duplicate pattern (should fail)..."
./ipam pattern create --name dev-gke-uswest --cidr-size 26 \
    --environment dev --region us-west1 \
    --block 10.0.0.0/8 --file default
print_result $? true

# Attempt to create a pattern with an invalid CIDR size (should fail)
echo "TEST: Creating pattern with invalid CIDR size (should fail)..."
./ipam pattern create --name invalid-cidr-size --cidr-size 33 \
    --environment dev --region us-west1 \
    --block 10.0.0.0/8 --file default
print_result $? true

# Attempt to create a pattern with a non-existent block (should fail)
echo "TEST: Creating pattern with non-existent block (should fail)..."
./ipam pattern create --name non-existent-block --cidr-size 26 \
    --environment dev --region us-west1 \
    --block 192.168.0.0/16 --file default
print_result $? true

# List all patterns for a block file
echo "TEST: Listing all patterns..."
./ipam pattern list --file default
print_result $? false

# Show details of an existing pattern
echo "TEST: Showing pattern details..."
./ipam pattern show --name dev-gke-uswest --file default
print_result $? false

# Attempt to show details of a non-existent pattern (should fail)
echo "TEST: Showing non-existent pattern details (should fail)..."
./ipam pattern show --name non-existent-pattern --file default
print_result $? true

# Create a subnet using a valid pattern
echo "TEST: Creating subnet using pattern..."
./ipam subnet create-from-pattern --pattern dev-gke-uswest \
    --file default
print_result $? false

# Create some subnets to test availability
echo "TEST: Creating test subnets..."
for i in {1..5}; do
    ./ipam subnet create --block 10.0.0.0/8 \
        --cidr 10.0.3.$((i*4))/30 \
        --name subnet-$i \
        --region us-west1
done

# Attempt to create a subnet using a non-existent pattern (should fail)
echo "TEST: Creating subnet using non-existent pattern (should fail)..."
./ipam subnet create-from-pattern --pattern non-existent-pattern --file default
print_result $? true

# Attempt to create a subnet when no available CIDR is left in the block (should fail)
echo "TEST: Creating subnet with no available CIDR (should fail)..."
./ipam subnet create-from-pattern --pattern dev-gke-uswest --file default
print_result $? true

# Clean up
final_cleanup

# Print overall test result
if [ "$all_tests_passed" = true ]; then
    echo -e "\033[32mALL TESTS PASSED\033[0m"
else
    echo -e "\033[31mSOME TESTS FAILED\033[0m"
fi