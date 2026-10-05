BINARY_NAME = quadrato
INSTALL_DIR = $(HOME)/bin
BUILD_DIR   = ./bin

.PHONY: build install uninstall clean

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "Make sure $(INSTALL_DIR) is in your PATH"

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)
