build:
	go build -o bin/main main.go

run:
	go run main.go

benchmark:
	go test -run=Bench -bench=.