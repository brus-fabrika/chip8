test:
	go test ./... -cover -coverprofile=c.out
	go tool cover -html=c.out

run:
	go build -o .\bin\chip8.exe && .\bin\chip8.exe
