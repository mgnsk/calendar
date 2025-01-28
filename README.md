## Simple calendar app

### Building and running locally.

You need the [Go](https://go.dev/) compiler.

- `$ make` - build the binary.
- `$ ./calendar --addr=:8080 --database-path=database.sqlite` - run the service.

Open your browser at http://localhost:8080

### Building with Docker Compose.

You need a Docker installation with the [compose plugin](https://docs.docker.com/compose/install/linux/).

- `$ docker compose build` - build the image.
- `$ docker compose up` - run the service.
