#!/bin/bash

# NOTE: Change makefile in accordance to what is being tested, Tsan/Asan, Valgrind or Helgrind
# After changing makefile, do "make clean" and "make" before running the bash script

ENGINE="../assignment1-e0725233_e0857381/engine"
GRADER="./grader"
#TEST_DIR="tests"
TEST_DIR="scripts"

# Loop through all .in files in the tests directory
for test_file in "$TEST_DIR"/*.in; do
    echo "Running test: $test_file"

    # For Tsan and Asan 
    $GRADER $ENGINE < "$test_file"


    # NOTE: Valgrind and Helgrind currently does not work
    # For Valgrind, note for valgrind we run it without the grader
    #valgrind --leak-check=full --show-leak-kinds=all --track-origins=yes $GRADER $ENGINE < "$test_file"

    # For Helgrind
    #valgrind --tool=helgrind $GRADER $ENGINE < "$test_file"

    echo "-------------------------------------"
done
