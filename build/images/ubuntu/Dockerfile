ARG RELEASE

FROM ubuntu:${RELEASE}

WORKDIR /workspace

COPY build/images/ubuntu/install.sh /workspace/bin/
COPY scripts/install-prerequisites-*.sh /workspace/bin/

RUN ./bin/install.sh
