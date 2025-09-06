BLOCKS_DIR := blocks
METADATA_FILE := metadata.json

.PHONY: cleanup
cleanup:
	@echo "Removing $(BLOCKS_DIR) and $(METADATA_FILE)"
	@rm -rf "$(BLOCKS_DIR)" "$(METADATA_FILE)"