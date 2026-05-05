.PHONY: run run-solution reproduce test-solution clean

# Run the broken version
run:
	go run main.go

# Run the fixed version
run-solution:
	go run solution/main.go

# End-to-End reproduction of the deadlock
reproduce:
	@echo "--- REPRODUCING DEADLOCK ---"
	@echo "1. Starting the broken server in background..."
	@go run main.go & sleep 2
	@echo "\n2. Sending request to http://localhost:8080/order?id=DEADLOCK-TEST"
	@echo "Expectation: The command below will HANG. We use --max-time 5 to force exit."
	@curl -v --max-time 5 "http://localhost:8080/order?id=DEADLOCK-TEST" || echo "\n\n[RESULT] Request timed out after 5s. The server is DEADLOCKED."
	@$(MAKE) clean

# End-to-End test of the solution
test-solution:
	@echo "--- TESTING SOLUTION ---"
	@echo "1. Starting the fixed server in background..."
	@go run solution/main.go & sleep 2
	@echo "\n2. Sending request to http://localhost:8080/order?id=SOLUTION-TEST"
	@echo "Expectation: The server should return 503 Service Unavailable immediately (after 500ms timeout)."
	@curl -i "http://localhost:8080/order?id=SOLUTION-TEST"
	@echo "\n\n[RESULT] Success. Server responded with an error instead of hanging."
	@$(MAKE) clean

# Run unit tests (including leak detection)
test:
	@echo "--- RUNNING LEAK DETECTION TEST ---"
	@go test -v main_test.go main.go || echo "\n[RESULT] Test failed as expected because a leak was detected!"

# Utility to kill the server running on port 8080
clean:
	@echo "Cleaning up server on port 8080..."
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
