FROM --platform=$BUILDPLATFORM alpine:3 AS certificates
RUN apk add --no-cache ca-certificates

FROM scratch
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
VOLUME /var/lib/cacao
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY cacao-${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /usr/bin/cacao
ENTRYPOINT ["/usr/bin/cacao"]
CMD ["--storage=/var/lib/cacao"]
