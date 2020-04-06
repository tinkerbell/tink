FROM ubuntu:xenial
COPY functions.sh /
COPY requirements.txt /tmp/osie/
COPY get-package-list.sh git-lfs-linux-*-v2.5.1-4-g2f166e02 /tmp/osie/
RUN apt-get update -y && \
    apt-get install -y $(/tmp/osie/get-package-list.sh) && \
    apt-get -qy autoremove && \
    apt-get -qy clean

RUN apt-get -qy update && \
    apt-get install -y \
      build-essential \
      libxml2-dev \
      libxslt1-dev \
      python3-dev \
      python3-pip \
      python3-setuptools \
      python3-wheel \
      systemd \
      zlib1g-dev && \
    pip3 install -r /tmp/osie/requirements.txt && \
    pip3 install --upgrade setuptools && \
    apt-get remove -y \
      build-essential \
      libxml2-dev \
      libxslt1-dev \
      python3-dev \
      python3-pip \
      python3-setuptools \
      python3-wheel \
      zlib1g-dev && \
    apt-get -qy autoremove && \
    apt-get -qy clean && \
    rm -rf /var/lib/apt/lists/* /tmp/osie ~/.cache/pip*
