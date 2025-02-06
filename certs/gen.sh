#!/usr/bin/env bash

set -e

openssl req -x509 -new -nodes -newkey rsa:2048 -keyout calendarCA.key \
	-sha256 -days 1825 -out calendarCA.crt -subj /CN='calendar.testing ca'

openssl req -newkey rsa:2048 -nodes -keyout calendar.testing.key -out calendar.testing.csr \
	-subj /CN=calendar.testing -addext subjectAltName=DNS:calendar.testing

openssl x509 -req -in calendar.testing.csr -copy_extensions copy \
	-CA calendarCA.crt -CAkey calendarCA.key -CAcreateserial -out calendar.testing.crt -days 365 -sha256
