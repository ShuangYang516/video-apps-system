FROM node:10.15.1-alpine as builder
WORKDIR /opt/web
COPY web .
RUN yarn install --ignore-optional && npm run build

FROM nginx:1.13.12

COPY --from=builder /opt/web/build /opt/cbvas
COPY docker/web/nginx-web.conf /etc/nginx/conf.d/default.conf
