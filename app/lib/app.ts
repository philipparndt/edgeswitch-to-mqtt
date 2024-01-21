import cron from "node-cron"
import { getAppConfig, Port } from "./config/config"
import { status, statusTotalTransmit, StatusType, turnOff, turnOn } from "./device/edge-switch"
import { log } from "./logger"

import { connectMqtt, mqttEmitter, publish } from "./mqtt/mqtt-client"

const portByName: any = {}

const statusUpdate = async (statusResult: StatusType, port: Port) => {
    if (statusResult) {
        publish(statusResult, `${port.name}/poe`)
    }
    else {
        publish({ message: "no response" }, `${port.name}/status`)
    }
}

export const triggerFullUpdate = async (config = getAppConfig().edgeswitch) => {
    let retry = 0
    while (retry < 3) {
        retry++
        try {
            const total = await statusTotalTransmit()
            let energySum = 0
            let whrSum = 0
            for (const port of config.ports) {
                const statusResult = await status(port.port)
                await statusUpdate(statusResult, port)

                const data = total[port.port]
                if (data) {
                    publish(data, `${port.name}/transmit`)
                }

                if (statusResult?.energy) {
                    energySum += statusResult.energy
                    whrSum += statusResult.total_Whr
                }
            }

            energySum = Math.round(energySum * 100) / 100
            whrSum = Math.round(whrSum * 100) / 100
            publish({ energySum, whrSum }, "aggregated")
        }
        catch (e) {
            log.error("Error while full update", e)
            if (retry < 3) {
                log.info("Retrying in 1 seconds.")
                await sleep(1_000)
            }
            else {
                throw e
            }
        }
    }

    log.debug("Full update done.")
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
        const statusResult = await status(port.port)
        await statusUpdate(statusResult, port)
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
        task.stop()
    }
}
