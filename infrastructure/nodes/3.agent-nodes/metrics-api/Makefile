.PHONY: run build execute-binary clean

BINARY=metrics-api

run:
	go run main.go

build:
	go build -o $(BINARY) main.go

execute-binary:
	$(BINARY)

clean:
	rm -f $(BINARY)
