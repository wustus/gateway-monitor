FROM golang:1.27.1 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev

RUN CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-X github.com/wustus/gateway-monitor/internal/version/version.Version=${VERSION}" \
  -o /gateway-monitor \
  .

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build /gateway-monitor /gateway-monitor

ENTRYPOINT ["/gateway-monitor"]
