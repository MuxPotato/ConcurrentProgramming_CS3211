#!/bin/bash

ENGINE="../cs3211-assignment-2-a2_e0725233_e0857381/engine"
GRADER="./grader"
#TEST_DIR="tests"
TEST_DIR="scripts"

# Loop through all .in files in the tests directory
for test_file in "$TEST_DIR"/*.in; do
    echo "Running test: $test_file"
    $GRADER $ENGINE < "$test_file"
    echo "-------------------------------------"
done