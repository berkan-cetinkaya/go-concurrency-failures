.PHONY: test-01 test-02 clean-all

# Experiment 01: Blocked Send
test-01:
	@$(MAKE) -C 01-blocked-send test

reproduce-01:
	@$(MAKE) -C 01-blocked-send reproduce

# Experiment 02: Missing Receiver
test-02:
	@$(MAKE) -C 02-missing-receiver test

reproduce-02:
	@$(MAKE) -C 02-missing-receiver reproduce

# Experiment 03: Worker Startup Failure
test-03:
	@$(MAKE) -C 03-worker-startup-failure test

reproduce-03:
	@$(MAKE) -C 03-worker-startup-failure reproduce

# Global cleanup
clean-all:
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@echo "All processes on port 8080 cleaned up."
