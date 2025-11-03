.PHONY: help seed test clean

help:
	@echo "Available commands:"
	@echo "  make seed-small    - Seed 100 test records"
	@echo "  make seed          - Seed 1,000 records (default)"
	@echo "  make seed-stress   - Seed 1,000,000 records (stress test)"
	@echo "  make seed-custom   - Seed with custom parameters (use RECORDS=n BATCH=n)"
	@echo "  make test          - Run filter performance tests"
	@echo "  make test-page     - Run tests with custom page (use PAGE=n SIZE=n)"
	@echo "  make clean         - Remove database file"
	@echo ""
	@echo "Examples:"
	@echo "  make seed-custom RECORDS=50000 BATCH=1000"
	@echo "  make test-page PAGE=2 SIZE=100"

seed-small:
	go run cmd/seed/main.go -preset=small

seed:
	go run cmd/seed/main.go -preset=default

seed-stress:
	go run cmd/seed/main.go -preset=stress

seed-custom:
	go run cmd/seed/main.go -count=$(or $(RECORDS),1000) -batch=$(or $(BATCH),500) -clear

test:
	go run cmd/test/main.go

test-page:
	go run cmd/test/main.go -page=$(or $(PAGE),1) -page-size=$(or $(SIZE),50)

clean:
	rm -f test.db

# Quick test workflow
quick: clean seed-small test

# Full stress test workflow
stress: clean seed-stress test
