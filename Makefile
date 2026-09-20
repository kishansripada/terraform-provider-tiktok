.PHONY: test testacc coverage check docs

test:
	go test ./...

testacc:
	TF_ACC=1 go test -race -run TestAcc ./...

coverage:
	TF_ACC=1 go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

check:
	test -z "$$(gofmt -l .)"
	go vet ./...
	go test ./...
	TF_ACC=1 go test -race ./...

docs:
	python3 scripts/generate-docs.py
