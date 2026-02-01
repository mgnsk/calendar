FROM node:25-bookworm@sha256:e6b32434aba48dcb8730d56de2df3d137de213f1f527a922a6bf7b2853a24e86 AS assets

WORKDIR /build

COPY package.json package-lock.json ./
ENV NODE_ENV=production
RUN npm ci

COPY tailwind.css ./
COPY html ./html
RUN npx @tailwindcss/cli -i tailwind.css -o app.css --minify


FROM golang:1.25.6-bookworm@sha256:2f768d462dbffbb0f0b3a5171009f162945b086f326e0b2a8fd5d29c3219ff14 AS build

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
