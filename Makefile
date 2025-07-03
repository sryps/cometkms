all: build

build:
	@echo "Building the binary"
	# Add your build commands here, e.g., compiling source files
	go build -o cometkms .

install:
	@echo "Installing the binary to GOBIN..."
	go install .

clean:
	@echo "Cleaning up build artifacts"
	rm -f cometkms


