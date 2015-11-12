PORT ?= 8000

.PHONY: gin
gin:
	go get github.com/codegangsta/gin
	ln -sf ../../static ./cmd/devhub/static
	ln -sf ../../bower_components ./cmd/devhub/bower_components
	cd ./cmd/devhub; gin --immediate --port=$(PORT) ./main.go
