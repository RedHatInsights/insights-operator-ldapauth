# Copyright 2020 Red Hat, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.13 AS builder

WORKDIR /insights-operator-ldapauth

COPY . /insights-operator-ldapauth

RUN make build

FROM registry.access.redhat.com/ubi8-minimal

COPY --from=builder /insights-operator-ldapauth/insights-operator-ldapauth .

RUN curl -L https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 -o /usr/bin/dumb-init && \
    curl -ksL https://url.corp.redhat.com/98bcda0 -o /etc/pki/ca-trust/source/anchors/RH-IT-Root-CA.crt && \
    update-ca-trust && \
    chmod a+x /usr/bin/dumb-init && \
    chmod a+x /insights-operator-ldapauth

USER 1001

ENTRYPOINT ["dumb-init", "--"]

CMD ["/insights-operator-ldapauth"]
