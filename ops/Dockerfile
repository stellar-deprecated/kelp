FROM ubuntu:20.04

LABEL maintainer="Nikhil Saraf <Github: @nikhilsaraf>"

# add dependencies: curl, unzip
RUN apt-get update && apt-get install -y curl unzip

# fetch ccxt-rest
RUN mkdir -p /root/.kelp/ccxt
RUN curl -L https://github.com/stellar/kelp/releases/download/ccxt-rest_v0.0.4/ccxt-rest_linux-x64.zip -o /root/.kelp/ccxt/ccxt-rest_linux-x64.zip
RUN unzip /root/.kelp/ccxt/ccxt-rest_linux-x64.zip -d /root/.kelp/ccxt

# fetch kelp archive
RUN curl -L https://github.com/stellar/kelp/releases/download/v1.12.0/kelp-v1.12.0-linux-amd64.tar -o kelp-archive.tar
# extract
RUN tar xvf kelp-archive.tar
# set working directory
WORKDIR kelp-v1.12.0

# set ulimit
RUN ulimit -n 10000

# start ccxt-rest when container is started
CMD nohup /root/.kelp/ccxt/ccxt-rest_linux-x64/ccxt-rest > ~/ccxt-rest.log &

# use command line arguments from invocation of docker run against this ENTRYPOINT command - https://stackoverflow.com/a/40312311/1484710
ENTRYPOINT ["./kelp"]
# default arguments are empty
CMD [""]

# sample command to run this container:
#     docker run nikhilsaraf/kelp:latest trade -c sample_trader.cfg -s buysell -f sample_buysell.cfg --sim
