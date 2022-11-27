import EventSource from "eventsource"
import cron from "node-cron"
import { getAppConfig } from "./config/config"
import { status } from "./device/edge-switch"
import { log } from "./logger"

import { connectMqtt, publish } from "./mqtt/mqtt-client"

let eventSource: EventSource

export const triggerFullUpdate = async (config = getAppConfig().edgeswitch) => {
    for (const port of config.ports) {
        const statusResult = await status(port)
        if (statusResult) {
            publish(statusResult, `${port}/status`)
        }
        else {
            publish({ message: "no response" }, `${port}/status`)
        }
    }
}

const start = async () => {
    // const ip = getAppConfig().denon.ip
    // console.log(`Connecting to Denon device on ${ip}`)
    // denonClient = new Denon.DenonClient(ip)
    // await denonClient.connect()
    //
    // denonClient.on("masterVolumeChanged", (volume: any) => {
    //     state.volume = volume
    //     publishState()
    // })
    //
    // denonClient.on("powerChanged", (power: any) => {
    //     state.power = power
    //     publishState()
    // })
    //
    // denonClient.on("error", async (error: any) => {
    //     console.log(error)
    //     state.power = "ERROR"
    //     publishState()
    //
    //     await start()
    // })

    await triggerFullUpdate()
}

export const startApp = async () => {
    const mqttCleanUp = await connectMqtt()
    await start()
    await triggerFullUpdate()
    log.info("Application is now ready.")

    log.info("Scheduling refresh.")
    const task = cron.schedule("* * * * *", () => triggerFullUpdate())
    task.start()

    return () => {
        mqttCleanUp()
        eventSource?.close()
        task.stop()
    }
}
