# syntax=docker/dockerfile:1
FROM golang:1.22-bullseye AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/logos_agency .

FROM gcr.io/distroless/static-debian11
COPY --from=builder /out/logos_agency /usr/local/bin/logos_agency
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/logos_agency"]
CMD []
