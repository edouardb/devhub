PORT ?= 8000

.PHONY: gin
gin:
	go get ./...
	go get github.com/codegangsta/gin
	ln -sf ../../static ./cmd/devhub/
	ln -sf ../../bower_components ./cmd/devhub/
	cd ./cmd/devhub; gin --immediate --port=$(PORT) ./main.go
