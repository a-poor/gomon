.PHONY: build
build: gomon.go
	go build -o build/gomon gomon.go

clean:
	rm -rf build/*
	rm -rf dist/*
	