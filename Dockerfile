FROM node:25-bookworm@sha256:508d817af6ec83e38d24a3d7da7da34b7119085c1d213abf0b7a3fab5dac5bf1 AS assets

WORKDIR /build

COPY package.json package-lock.json ./
ENV NODE_ENV=production
RUN npm ci

COPY tailwind.css ./
COPY html ./html
RUN npx @tailwindcss/cli -i tailwind.css -o app.css --minify


FROM golang:1.26.1-bookworm@sha256:c7a82e9e2df2fea5d8cb62a16aa6f796d2b2ed81ccad4ddd2bc9f0d22936c3f2 AS build

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
