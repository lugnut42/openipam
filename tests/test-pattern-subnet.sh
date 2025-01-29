#!/bin/bash

source ./tests/test-support.sh

# Initialize test environment
init_test_env

echo "TEST: Adding block..."
./ipam block create --cidr 10.0.0.0/8 --file default --description "Test block"
print_result $? false

echo "TEST: Creating pattern..."
./ipam pattern create \
    --name test-pattern \
    --cidr-size 24 \
    --environment dev \
    --region us-west1 \
    --block 10.0.0.0/8 \
    --file default
print_result $? false

echo "TEST: Creating first subnet from pattern..."
./ipam subnet create-from-pattern \
    --pattern test-pattern \
    --file default
print_result $? false

echo "TEST: Verifying subnet was created..."
./ipam subnet list --block 10.0.0.0/8
print_result $? false

echo "TEST: Creating second subnet from pattern..."
./ipam subnet create-from-pattern \
    --pattern test-pattern \
    --file default
print_result $? false

echo "TEST: Verifying both subnets exist..."
./ipam subnet list --block 10.0.0.0/8
print_result $? false

# Clean up
final_cleanup