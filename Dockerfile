FROM golang:1.23-bookworm AS build

WORKDIR /build

COPY . .

RUN make


FROM gcr.io/distroless/base-debian12

COPY --from=build /build/calendar /
