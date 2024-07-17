FROM golang:1.22.5-alpine AS build

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0
RUN go mod tidy
RUN go build -o /app/bagel -ldflags="-s -w" /app/main.go

FROM alpine:3.20

RUN apk update
RUN apk add --no-cache unzip python3 py3-pip

RUN addgroup -S bagel 
RUN adduser -S bagel -G bagel -h /home/bagel
USER bagel
WORKDIR /home/bagel

RUN pip install semgrep --user --break-system-packages
COPY --from=build /app/bagel .

ENV INSIDETHEMATRIX=true
ENV PATH="/home/bagel/.local/bin:$PATH" 
EXPOSE 8080

ENTRYPOINT ["./bagel"]