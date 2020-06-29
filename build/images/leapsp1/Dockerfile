FROM opensuse/leap:15.1

WORKDIR /workspace

COPY build/images/sle15sp1/install.sh /workspace/bin/
COPY scripts/install-prerequisites-*.sh /workspace/bin/

RUN ./bin/install.sh
