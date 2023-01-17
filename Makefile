
all: websvc

websvc:
	go build -o bin ./cmd/$@

test:
	set PG_DSN=$(PG_DSN)
	go test ./...
