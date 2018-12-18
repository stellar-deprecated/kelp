### STAGE 1: Build ###

FROM node:8-alpine AS vue-built
COPY . /

# we need that for node-gyp
RUN apk add --update \
    autoconf \
    automake

RUN npm set progress=false && npm config set depth 0

## Storing node modules on a separate layer will prevent unnecessary npm installs at each build
RUN npm i node-sass
RUN npm i

## Build the vue app in production mode and store the artifacts in dist folder
RUN npm run build

### STAGE 2: Setup ###

FROM nginx:1.15.2-alpine

## Copy our default nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

## Remove default nginx website
RUN rm -rf /usr/share/nginx/html/*

## From 'builder' stage copy over the artifacts in dist folder to default nginx public folder
COPY --from=vue-built /build /usr/share/nginx/html

## copy favicons
# COPY --from=vue-built /favicon/* /usr/share/nginx/html/

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
