#!/bin/bash
ENGINE="../cs3211-assignment-2-a2_e0725233_e0857381/engine"
GRADER="./grader"
#TEST_DIR="tests"
TEST_DIR="scripts"
TEST_FILE="$TEST_DIR/test11.in"
REPEAT=20

if [ ! -f "$TEST_FILE" ]; then
    echo "Test file not found: $TEST_FILE"
    exit 1
fi

pass_count=0
echo "Running test: $TEST_FILE"
for i in $(seq 1 $REPEAT); do
    echo -n "  Run $i: "
    $GRADER $ENGINE < "$TEST_FILE"
    STATUS=$?
    if [[ $STATUS -ne 0 ]]; then
        echo "FAILED (exit code $STATUS)"
    else
        echo "PASSED"
        ((pass_count++))
    fi
done

echo "-------------------------------------"
echo "✅ $pass_count/$REPEAT runs passed"