FROM golang:1.13.7 AS build

ARG ORACLE_VERSION
ENV ORACLE_VERSION=${ORACLE_VERSION}
ENV LD_LIBRARY_PATH "/usr/lib/oracle/${ORACLE_VERSION}/client64/lib"

RUN apt-get -qq update && apt-get install --no-install-recommends -qq libaio1 rpm
COPY oci8.pc.template /usr/share/pkgconfig/oci8.pc
RUN sed -i "s/@ORACLE_VERSION@/$ORACLE_VERSION/g" /usr/share/pkgconfig/oci8.pc
COPY oracle*${ORACLE_VERSION}*.rpm /
COPY instantclient-basic-linux.x64-19.5.0.0.0dbru.zip /
RUN rpm -Uh --nodeps /oracle-instantclient*.x86_64.rpm && rm /*.rpm
RUN echo $LD_LIBRARY_PATH >> /etc/ld.so.conf.d/oracle.conf && ldconfig

WORKDIR /go/src/prometheus_oracle
COPY . .
RUN go get -d -v

ARG VERSION
ENV VERSION ${VERSION:-0.1.0}

ENV PKG_CONFIG_PATH /go/src/prometheus_oracle
ENV GOOS            linux

RUN go build -v -ldflags "-X main.Version=${VERSION} -s -w"

# new stage
FROM frolvlad/alpine-glibc
LABEL authors="Denis Evsyukov"
LABEL maintainer="Denis Evsyukov <denis@evsyukov.org>"

ENV VERSION ${VERSION:-0.1.0}

COPY instantclient-basic-linux.x64-19.5.0.0.0dbru.zip /

RUN apk add --no-cache libaio unzip && \
    unzip /instantclient-basic-linux.x64-19.5.0.0.0dbru.zip && \
    rm -f /instantclient-basic-linux.x64-19.5.0.0.0dbru.zip && \
    apk del unzip

ARG ORACLE_VERSION
ENV ORACLE_VERSION=${ORACLE_VERSION}
ENV LD_LIBRARY_PATH "/instantclient_19_5"

COPY --from=build /go/src/prometheus_oracle/prometheus_oracle /prometheus_oracle

RUN chmod 755 /prometheus_oracle

EXPOSE 9101

ENTRYPOINT ["/prometheus_oracle"]
