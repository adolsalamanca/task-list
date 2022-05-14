.phony: run test

run:
	@go build -o task-list && ./task-list

test:
	@go test ./... -coverprofile=c.out -covermode=atomic -race

coverage:
	@go tool cover -html=c.out

clean:
	@rm -f task-list