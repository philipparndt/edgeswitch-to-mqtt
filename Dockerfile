FROM node:19.9-alpine
COPY app/dist /opt/app/
WORKDIR /opt/app/

CMD ["node", "index.js", "/var/lib/edgeswitch-to-mqtt-gw/config.json"]
