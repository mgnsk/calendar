## Simple calendar app

### Local development

- `make setup` - Set up the development sandbox.
- `make dev` - Start the development sandbox.
- `make stop` - Stop the development sandbox.

The calendar application can be visited on `https://calendar.localhost`.

### Production installation

The Docker image is published at https://github.com/mgnsk/calendar/pkgs/container/calendar

#### Clone the repository

- `mkdir /opt/mgnsk/calendar`
- `cd /opt/mgnsk/calendar`
- `git clone https://github.com/mgnsk/calendar.git .`

#### Create the `.env` file with your public domain name

```
HOSTNAME="my-domain-name.com"
```

#### Install and enable the service

- `make install`
