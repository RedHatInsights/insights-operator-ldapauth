FROM golang:1.13 AS builder

WORKDIR /insights-operator-ldapauth

COPY . /insights-operator-ldapauth

RUN make build

FROM registry.access.redhat.com/ubi8-minimal

COPY --from=builder /insights-operator-ldapauth/insights-operator-ldapauth .

RUN curl -L https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 -o /usr/bin/dumb-init && \
    chmod a+x /usr/bin/dumb-init && \
    chmod a+x /insights-operator-ldapauth

USER 1001

ENTRYPOINT ["dumb-init"]

CMD ["/insights-operator-ldapauth"]
