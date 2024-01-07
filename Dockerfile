FROM node:20.5
COPY app/dist /opt/app/
WORKDIR /opt/app/

CMD ["node", "index.js", "/var/lib/edgeswitch-to-mqtt-gw/config.json"]
