FROM scratch
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
VOLUME /var/lib/cacao
COPY cacao-${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /usr/bin/cacao
ENTRYPOINT ["/usr/bin/cacao"]
CMD ["--storage=/var/lib/cacao"]
