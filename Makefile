BINARY       = bin/shai
BINARY_DEBUG = bin/shai-dbg

build:
	go build -o $(BINARY) .

debug:
	go build -tags debug -o $(BINARY_DEBUG) .

all: build debug

clean:
	rm -rf bin/