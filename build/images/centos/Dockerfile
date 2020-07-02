ARG RELEASE

FROM centos:${RELEASE}

WORKDIR /workspace

# Shadow the bogus /etc/resolv.conf of centos:8 by copying a blank file over it
COPY build/images/centos/resolv.conf /etc/

COPY build/images/centos/install.sh /workspace/bin/
COPY scripts/install-prerequisites-*.sh /workspace/bin/

RUN ./bin/install.sh
