FROM golang:1.23-bookworm AS deps

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download


FROM deps AS build

COPY . .

RUN make


FROM gcr.io/distroless/base-debian12

COPY --from=build /build/calendar /
