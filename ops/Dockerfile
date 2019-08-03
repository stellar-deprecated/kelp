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
RUN curl -s https://glide.sh/get | sh

# install kelp
ENV GOPATH /go
RUN git clone https://github.com/stellar/kelp.git .
RUN git fetch --tags
RUN glide install
RUN ./scripts/build.sh

# set ulimit
RUN ulimit -n 10000

# use command line arguments from invocation of docker run against this ENTRYPOINT command - https://stackoverflow.com/a/40312311/1484710
ENTRYPOINT ["./bin/kelp"]

# sample command to run this container as a daemon process:
# docker run -v `pwd`/ops:/go/src/github.com/stellar/kelp/bin/ops -d -p 8011:8011 <image> server -p 8011 --ccxt-rest-url=http://host.docker.internal:3000
# this assumes that you are running ccxt on port 3000 outside this kelp docker container
# the three port numbers (8011 in the example above) must be the same and must be specified
# the first part of the -v argument is the directory where you want to save the kelp configs and kelp logs from the bots created in this container
