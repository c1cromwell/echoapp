# Echo overlay of Euclid's infra/metagraph-ubuntu/Dockerfile (copied over the checkout by
# metagraph/scripts/setup-euclid.sh). Difference from upstream: the Tessellation node binaries
# come from Echo's own vendored, checksum-verified copies, or are built from Echo's fork. Nothing
# here contacts Constellation's GitHub. See docs/OWN_L0_PLAN.md §9 stage 1.
FROM ubuntu:20.04

WORKDIR /code

ARG CHECKOUT_TESSELLATION_VERSION
ARG TESSELLATION_VERSION_IS_TAG_OR_BRANCH
ARG TESSELLATION_REPO_URL=https://github.com/c1cromwell/tessellation.git

ENV LC_ALL=C.UTF-8 \
    LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8 \
    DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      openjdk-21-jre curl gnupg lsof ca-certificates git && \
    echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | tee /etc/apt/sources.list.d/sbt.list && \
    echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | tee /etc/apt/sources.list.d/sbt_old.list && \
    curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | apt-key add - && \
    apt-get update && \
    apt-get install -y sbt && \
    rm -rf /var/lib/apt/lists/*

# vendor/<version>/ holds cl-node.jar, cl-dag-l1.jar, cl-wallet.jar, cl-keytool.jar and SHA256SUMS.
COPY vendor/ /code/vendor/

RUN set -eux; \
    V="/code/vendor/${CHECKOUT_TESSELLATION_VERSION}"; \
    if [ -f "$V/cl-node.jar" ]; then \
        (cd "$V" && sha256sum --ignore-missing -c SHA256SUMS); \
        cp "$V/cl-wallet.jar" cl-wallet.jar; \
        cp "$V/cl-keytool.jar" cl-keytool.jar; \
        cp "$V/cl-dag-l1.jar" dag-l1.jar; \
        cp "$V/cl-node.jar" global-l0.jar; \
    else \
        git clone "$TESSELLATION_REPO_URL" tessellation; \
        cd tessellation; \
        git checkout "$CHECKOUT_TESSELLATION_VERSION" || exit 1; \
        if [ "$TESSELLATION_VERSION_IS_TAG_OR_BRANCH" = "branch" ]; then \
        export RELEASE_TAG=99.99.99; \
        sbt -Dversion.override=99.99.99 sdk/publishM2; \
        fi; \
        sbt wallet/assembly dagL1/assembly keytool/assembly dagL0/assembly; \
        mv modules/wallet/target/scala-2.13/tessellation-wallet-assembly-* /code/cl-wallet.jar; \
        mv modules/keytool/target/scala-2.13/tessellation-keytool-assembly-* /code/cl-keytool.jar; \
        mv modules/dag-l1/target/scala-2.13/tessellation-dag-l1-assembly-* /code/dag-l1.jar; \
        mv modules/dag-l0/target/scala-2.13/tessellation-dag-l0-assembly-* /code/global-l0.jar; \
        cd /code && rm -rf tessellation; \
    fi; \
    rm -rf /code/vendor
