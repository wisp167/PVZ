# Builder stage
FROM golang:1.23.4 as builder

WORKDIR /app


ENV GO111MODULE=on
ENV GOPATH=/go

RUN apt-get update && apt-get install -y make
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
RUN go install github.com/air-verse/air@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.28.0

COPY go.mod go.sum Makefile ./
RUN go mod download

COPY . .

RUN mkdir -p /app/bin

RUN make clean gen
RUN go build -o /app/bin/pvz-service ./main.go

# Runner stage (with Air for development)
FROM golang:1.23.4 as runner

WORKDIR /app



COPY --from=builder /go/bin/air /go/bin/
COPY --from=builder /app/bin/pvz-service ./bin/
COPY --from=builder /app/.air.toml .
COPY --from=builder /app/go.mod .
COPY --from=builder /app/go.sum .
COPY --from=builder /app/main.go .
RUN go mod download

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM golang:1.23.4 as tester

WORKDIR /app

COPY --from=builder /app .

RUN go mod download

RUN ls -la api/
RUN go list -f '{{.GoFiles}}' github.com/wisp167/pvz/api

CMD ["go", "test", "-v", "./tests/..."]

# Test stage
#FROM builder as tester
#COPY tests /app/tests
#CMD ["go", "test", "-v", "./...", "-coverprofile=coverage.out"]
