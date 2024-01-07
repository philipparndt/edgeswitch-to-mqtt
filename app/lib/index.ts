/* istanbul ignore file */
import * as path from "path"
import { startApp } from "./app"
import { loadConfig } from "./config/config"

import { log } from "./logger"

process.on("uncaughtException", (err: any) => {
    if (err.code === "ECONNRESET") {
        // ignore
    }
    else {
        log.error("Uncaught Exception thrown, Bye bye!", err)
        process.exit(-1)
    }
})

process.on("SIGINT", () => {
    log.info("SIGINT, Bye bye!")
    process.exit(0)
})

process.on("SIGTERM", () => {
    log.info("SIGTERM, Bye bye!")
    process.exit(0)
})

if (process.argv.length !== 3) {
    log.error("Expected config file as argument.")
    process.exit(1)
}

let configFile = process.argv[2]
configFile = configFile.startsWith(".") ? path.join(__dirname, "..", configFile) : configFile
log.info(`Using config from file ${configFile}`)

loadConfig(configFile)

startApp().then()
