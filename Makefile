install-tools:
	if [ ! $$(which go) ]; then \
		echo "goLang not found."; \
		echo "Try installing go..."; \
		exit 1; \
	fi
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.1
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/axw/gocov/gocov@latest
	go get golang.org/x/tools/cmd/goimports
	# apt install gcc
	go install github.com/AlekSi/gocov-xml@latest
	if [ ! $$( which migrate ) ]; then \
		echo "The 'migrate' command was not found in your path. You most likely need to add \$$HOME/go/bin to your PATH."; \
		exit 1; \
	fi

gen:
	docker build -t cliqkets-mockgen -f Dockerfile.mock .
	go mod tidy
	docker run  --name generat-mock-container --rm  -v $$(pwd):/tmp/mock-vol -w /tmp/mock-vol cliqkets-mockgen sh -c 'go generate ./...' 
fix:
	golangci-lint run --fix	

lint:fix
	golangci-lint run ./...

tidy:
	go mod tidy

test: tidy
	gocov test ./... | gocov report 