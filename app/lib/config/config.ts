import * as fs from "fs"
import { log } from "../logger"

export type ConfigMqtt = {
    url: string,
    topic: string
    username?: string
    password?: string
    retain: boolean
    qos: (0 | 1 | 2)
    "bridge-info"?: boolean
    "bridge-info-topic"?: string
}

export type Port = {
    name: string
    port: string
}

export type ConfigEdgeSwitch = {
    "ip": string
    "username": string
    "password": string

    "ports": Port[]
}

export type Config = {
    mqtt: ConfigMqtt
    edgeswitch: ConfigEdgeSwitch
    loglevel: string
}

let appConfig: Config

const configDefaults = {
    loglevel: "info"
}

const mqttDefaults = {
    qos: 1,
    retain: true,
    "bridge-info": true
}

export const applyDefaults = (config: any) => {
    return {
        ...configDefaults,
        ...config,
        mqtt: { ...mqttDefaults, ...config.mqtt }
    } as Config
}

export const loadConfig = (file: string) => {
    const buffer = fs.readFileSync(file)
    applyConfig(JSON.parse(buffer.toString()))
    return appConfig
}

export const applyConfig = (config: any) => {
    appConfig = applyDefaults(config)
    log.configure(appConfig.loglevel.toUpperCase())
}

export const getAppConfig = () => {
    return appConfig
}
