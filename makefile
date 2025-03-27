VERSION=$(shell git describe --tags 2>/dev/null || echo "snapshot")

build:
	go mod tidy
	env CGO_ENABLED=0 GOOS=linux go build -o bin/merge-raml -ldflags "-X main.version=${VERSION}" cmd/main.go

release:
	env CGO_ENABLED=0 GOOS=linux go build -o bin/merge-raml -ldflags "-X main.version=${VERSION}" cmd/main.go
	cd bin && tar -czf merge-raml-linux.${VERSION}.tar.gz merge-raml && rm merge-raml

	env CGO_ENABLED=0 GOOS=darwin go build -o bin/merge-raml -ldflags "-X main.version=${VERSION}" cmd/main.go
	cd bin && tar -czf merge-raml-darwin.${VERSION}.tar.gz merge-raml && rm merge-raml

	env CGO_ENABLED=0 GOOS=windows go build -o bin/merge-raml.exe -ldflags "-X main.version=${VERSION}" cmd/main.go
	cd bin && zip -9 merge-raml-windows.${VERSION}.zip merge-raml.exe && rm merge-raml.exe