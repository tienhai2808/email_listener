ROOT := .
TMP_DIR := tmp
BIN := $(TMP_DIR)/main
TESTDATA_DIR := testdata

.PHONY: build run clean

build:
	@echo "Building..."
	@mkdir -p $(TMP_DIR)
	go build -o $(BIN) .

run: build
	@echo "Running..."
	@$(BIN)

clean:
	@echo "Cleaning..."
	@rm -rf $(TMP_DIR)