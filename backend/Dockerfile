FROM golang:1.25.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/config-manager ./cmd/config-manager

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/config-manager /config-manager

EXPOSE 8080

ENTRYPOINT ["/config-manager"]

