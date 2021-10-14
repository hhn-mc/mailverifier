FROM golang AS builder
LABEL stage=intermediate
COPY . /mailverifier
WORKDIR /mailverifier/cmd/mailverifier
ENV GO111MODULE=on
RUN GOOS=linux GOARCH=amd64 go build -o /main .

FROM scratch
LABEL maintainer="Hendrik Jonas Schlehlein <hendrik.schlehlein@gmail.com>"
WORKDIR /
COPY --from=builder /main ./
ENTRYPOINT [ "./main" ]