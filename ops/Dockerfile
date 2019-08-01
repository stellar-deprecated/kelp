FROM golang:1.12.7

LABEL maintainer="Nikhil Saraf <Github: @nikhilsaraf>"

# install yarn
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt update
RUN apt -y install yarn

# set working dir
WORKDIR /go/src/github.com/stellar/kelp
RUN pwd

# install glide
RUN curl https://glide.sh/get | sh

# install kelp
ENV GOPATH /go
RUN git clone https://github.com/stellar/kelp.git .
RUN glide install
RUN ./scripts/build.sh

# set ulimit
RUN ulimit -n 10000

CMD ["./bin/kelp", "server", "-p", "8000", "--horizon-testnet-uri=https://horizon-dev-testnet.stellar.org"]
