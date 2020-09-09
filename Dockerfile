FROM golang:1.13.5

ADD . /go/src/github.com/dmitry-msk777/Connector_1C_Enterprise
WORKDIR /go/src/github.com/dmitry-msk777/Connector_1C_Enterprise

RUN go get "github.com/streadway/amqp"
RUN go get "github.com/gorilla/mux"

RUN CGO_ENABLED=0 go install github.com/dmitry-msk777/Connector_1C_Enterprise

ENTRYPOINT /go/bin/Connector_1C_Enterprise

EXPOSE 8181

# EXPOSE 5672 8181 15672

# https://temofeev.ru/info/articles/kak-sozdat-prostoy-mikroservis-na-golang-i-grpc-i-vypolnit-ego-konteynerizatsiyu-s-pomoshchyu-docker/
# https://habr.com/ru/post/238473/

# docker build -t connector_1c_enterprise .
# docker run --publish 18184:5300 --name test --rm connector_1c_enterprise

# docker run -p 5672:5672 -p 15672:15672 -p 8181:8181 --name test --rm connector_1c_enterprise
# docker run -p 8181:8181 --name microservice_1c --rm connector_1c_enterprise --link rabbitmq:rb

# https://hub.docker.com/repository/docker/dmitrymsk777/connector_1c_enterprise
# docker run dmitrymsk777/connector_1c_enterprise