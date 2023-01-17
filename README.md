
# Information
Web service binary code is located at `cmd` directory. Then internal and shared code are located in `src`.

Protobuf schema is located at `schema` directory.

Docker compose file is located at `docker` directory.

## Compile
This project include Makefile that can be used with `nmake` (Visual Studio NMake)
```shell
nmake
```

However, manual compile can also be done with these command:
```shell
go build -o bin ./cmd/websvc
```

Both command above will result in two binary in bin directory:
```shell
./bin/websvc.exe
```

## Running
### Windows (Powershell)
#### 1. Run postgresql
Using provided `docker/docker-compose.yml` file, we can run db on local machine and it will listen on `0.0.0.0:5432`.

```shell
cd docker
docker-compose up -d
```

#### 2. Running webapi
Add required configuration first by setting up env variables below:
```shell
$env:HTTP_PORT="127.0.0.1:9092"
$env:PG_DSN="postgresql://postgres:secretpassword123@localhost:5432/"
```

Please adjust based on your OS and shell. For example if you are on linux, you might need to call `export` instead. And
for windows `cmd` user, you will need to call `set`.

Then enter to `bin` directory and run `processor` executable.
```shell
cd bin
.\websvc.exe
```

## Testing
This project include unit test that can be executed using `nmake`
```shell
$env:PG_DSN="postgresql://postgres:secretpassword123@localhost:5432/"
nmake test
```

or you can run manually via
```shell
$env:PG_DSN="postgresql://postgres:secretpassword123@localhost:5432/"
go test ./...
```