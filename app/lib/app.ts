import EventSource from "eventsource"
import cron from "node-cron"
import { getAppConfig, Port } from "./config/config"
import { status, statusTotalTransmit, turnOff, turnOn } from "./device/edge-switch"
import { log } from "./logger"

import { connectMqtt, mqttEmitter, publish } from "./mqtt/mqtt-client"

let eventSource: EventSource

const portByName: any = {}

const statusUpdate = async (port: Port) => {
    const statusResult = await status(port.port)
    if (statusResult) {
        publish(statusResult, `${port.name}/poe`)
    }
    else {
        publish({ message: "no response" }, `${port.name}/status`)
    }
}

export const triggerFullUpdate = async (config = getAppConfig().edgeswitch) => {
    const total = await statusTotalTransmit()

    for (const port of config.ports) {
        await statusUpdate(port)

        const data = total[port.port]
        if (data) {
            publish(data, `${port.name}/transmit`)
        }
    }
}

const start = async () => {
    await triggerFullUpdate()
}

const getPort = (topic: string) => {
    const parts = topic.split("/")
    const port = parts.length > 3 ? parts[parts.length - 3] : undefined
    if (port) {
        return portByName[port]
    }
    return undefined
}

const sleep = (ms: number) => {
    return new Promise((resolve) => setTimeout(resolve, ms))
}

const onSet = async (topic: string, message: any) => {
    const port = getPort(topic)
    console.log("Received message", topic, message, port)
    if (port) {
        if (message === true) {
            await turnOn(port.port)
        }
        else if (message === false) {
            await turnOff(port.port)
        }

        await sleep(10_000)
        await statusUpdate(port)
    }
}

export const startApp = async () => {
    for (const port of getAppConfig().edgeswitch.ports) {
        portByName[port.name] = port
    }

    const mqttCleanUp = await connectMqtt()
    await start()
    await triggerFullUpdate()
    log.info("Application is now ready.")

    log.info("Scheduling refresh.")
    const task = cron.schedule("* * * * *", () => triggerFullUpdate())
    task.start()

    mqttEmitter.on("/set", async (data: any) => {
        const topic = data.topic
        const message = JSON.parse(data.message.toString())
        await onSet(topic, message)
    })

    return () => {
        mqttCleanUp()
        eventSource?.close()
        task.stop()
    }
}
