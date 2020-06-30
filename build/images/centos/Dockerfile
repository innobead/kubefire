ARG RELEASE

FROM centos:${RELEASE}

WORKDIR /workspace

COPY build/images/centos/install.sh /workspace/bin/
COPY scripts/install-prerequisites-*.sh /workspace/bin/

RUN ./bin/install.sh
