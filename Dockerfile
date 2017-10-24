FROM golang:1.9.1 as builder
WORKDIR /go/src/rhinoceros
COPY . .
RUN go get && go build -o rhinoceros .

FROM scratch
WORKDIR /root/
COPY --from=builder /go/src/rhinoceros/rhinoceros
CMD ["./rhinoceros"]
