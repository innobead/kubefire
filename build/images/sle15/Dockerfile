ARG RELEASE

FROM registry.suse.com/suse/sle15:${RELEASE}

WORKDIR /workspace

COPY build/images/sle15/install.sh /workspace/bin/

RUN ./bin/install.sh
