## Simple calendar app

### Local development

Dependencies:

- [Go](https://go.dev/)
- [Node.js](https://nodejs.org/en)

Add the following to your `/etc/hosts` to access the calendar by domain name.

```
127.0.0.1 calendar.testing
```

Generate the root CA and certificate:

```
cd certs
./gen.sh
```

During development, we use self-signed certificates. In production we use automatic TLS from Let's Encrypt.

Import Calendar CA certificate at `certs/ca.crt` into your browser to avoid self-signed certificate warnings.

Setup tools and run the development environment:

- `$ make setup`
- `$ make dev`

Your browser should automatically open at `https://calendar.testing:8443`.

### Docker Compose deployment

The Docker image is published at https://github.com/mgnsk/calendar/pkgs/container/calendar

Describes an easy deployment pattern I personally use elsewhere.
Choose whichever system you're familiar with.

You need a Docker installation with the [Compose plugin](https://docs.docker.com/compose/install/linux/).

Copy `docker-compose.example.yml` to `/etc/docker/compose/calendar/docker-compose.yml`.
Configure the environment variables.

Copy `docker-compose@.service` to `/etc/systemd/system/`.

Enable the service: `$ systemctl enable --now docker-compose@calendar`.

To update the service:

- `$ cd /etc/docker/compose/calendar`
- `$ docker compose pull`
- `$ systemctl restart docker-compose@calendar`

TODO: automatic update
