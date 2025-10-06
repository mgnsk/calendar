FROM node:24-bookworm@sha256:a2ed436bacdcc9dd543202a327bbce2519c43e3755a41a186f8f51c037ef3342 AS assets

WORKDIR /build

COPY package.json package-lock.json ./
ENV NODE_ENV=production
RUN npm ci

COPY tailwind.css ./
COPY html ./html
RUN npx @tailwindcss/cli -i tailwind.css -o app.css --minify


FROM golang:1.25.1-bookworm@sha256:2960a1db140a9a6dd42b15831ec6f8da0c880df98930411194cf11875d433021 AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
COPY --from=assets /build/app.css ./app.css
COPY --from=assets /build/node_modules ./node_modules

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOCACHE=/go/cache

RUN --mount=type=cache,target=/go/cache/ go build -trimpath -tags timetzdata,strictdist -o calendar ./cmd/calendar

RUN mkdir /database /cache


FROM scratch

COPY --from=build /build/calendar /bin/calendar
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chown=65534:65534 /cache/ /cache/
COPY --from=build --chown=65534:65534 /database/ /database/

USER 65534:65534
