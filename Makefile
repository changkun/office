# Copyright 2021 Changkun Ou. All rights reserved.
# Use of this source code is governed by a MIT
# license that can be found in the LICENSE file.

VERSION = $(shell git describe --always --tags)
IMAGE = officecheck
BINARY = officecheck
TARGET = -o $(BINARY)
BUILD_FLAGS = $(TARGET)

all:
	go build $(BUILD_FLAGS)
run:
	./$(BINARY)
clean:
	rm -rf $(BINARY)