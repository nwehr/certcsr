FROM golang:latest

WORKDIR /certcsr

COPY . .

RUN cd ./backend \
    && go get -d -v ./... \
    && go install -v ./...

CMD ["backend"]