BINARY       = bin/_shai_bin
BINARY_DEBUG = bin/_shai_bin_dbg

build:
	go build -o $(BINARY) .

debug:
	go build -tags debug -o $(BINARY_DEBUG) .

all: build debug

clean:
	rm -rf bin/