VERSION=$(shell git describe --tags 2>/dev/null || echo "snapshot")

build:
	env CGO_ENABLED=0 GOOS=linux go build -o bin/merge-raml -ldflags "-X main.version=${VERSION}" cmd/main.go

travis:
	env CGO_ENABLED=0 GOOS=linux go build -o bin/merge-raml -ldflags "-X main.version=${VERSION}" cmd/main.go
	cd bin && tar -czf merge-raml-linux.${TRAVIS_TAG}.tar.gz ./merge-raml && rm merge-raml

	env CGO_ENABLED=0 GOOS=darwin go build -o bin/merge-raml -ldflags "-X main.version=${VERSION}" cmd/main.go
	cd bin && tar -czf merge-raml-darwin.${TRAVIS_TAG}.tar.gz ./merge-raml && rm merge-raml

	env CGO_ENABLED=0 GOOS=windows go build -o bin/merge-raml.exe -ldflags "-X main.version=${VERSION}" cmd/main.go
	cd bin && zip -9 merge-raml-windows.${TRAVIS_TAG}.zip ./merge-raml.exe && rm merge-raml.exe