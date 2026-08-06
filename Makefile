BINARIES := nest-server nest-desktop nest-workstation

.PHONY: all $(BINARIES) test clean

all: $(BINARIES)

nest-server:
	go build -o bin/$@ ./cmd/nest-server

nest-desktop:
	go build -o bin/$@ ./cmd/nest-desktop

nest-workstation:
	go build -o bin/$@ ./cmd/nest-workstation

test:
	go test ./...

clean:
	rm -rf bin/