FROM golang:1.18 as build
WORKDIR /workspace
COPY go.mod .
COPY go.sum .
RUN go mod download
ADD . /workspace
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build ./cmd/main.go

FROM alpine:3.16
RUN adduser -D app
USER app
WORKDIR /home/app
COPY --from=build /workspace/main .
ENV MF_ID_PASSWORD ""
CMD ["./main"]
