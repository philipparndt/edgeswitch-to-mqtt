import AsyncLock from "async-lock"
import { NodeSSH } from "node-ssh"
import { getAppConfig } from "../config/config"
import { parseTotalTransmit } from "./total-transmit-parser"

const lock = new AsyncLock({ timeout: 5000 })

const exec = async (commands: string[], config = getAppConfig().edgeswitch): Promise<string> => {
    const messages: string[] = []

    await lock.acquire<string>("ssh", async (done) => {
        const ssh = new NodeSSH()

        const client = await ssh.connect({
            host: config.ip,
            username: config.username,
            password: config.password
        })

        return new Promise((resolve) => {
            client.requestShell().then(async (shell) => {
                shell.on("data", (data: Buffer) => {
                    messages.push(data.toString("utf8"))
                })
                shell.on("close", async () => {
                    resolve(messages.join(""))
                    done()
                })

                for (const command of commands) {
                    await shell.write(`${command}\n`)
                }
                await shell.end()
            })
        })
    })

    return messages.join("")
}

export const turnOff = async (port: string) => {
    await exec([
        "configure",
        `interface ${port}`,
        "poe opmode shutdown"]
    )
}

export const turnOn = async (port: string) => {
    await exec([
        "configure",
        `interface ${port}`,
        "poe opmode auto"]
    )
}

export const status = async (port: string) => {
    const message = await exec([
        "configure",
        `show poe status ${port}`]
    )

    const result = (message ?? "")
        .split("\n")
        .filter((line) => line.startsWith(port))
        .map(line => line.trim())
        .map((line) => line.split(/\s+/g))[0]

    if (!result || result.length === 0) {
        return undefined
    }

    return {
        interface: result[0],
        detection: result[1],
        status: result[1].toLowerCase() === "good" ? 1 : 0,
        class: result[2],
        energy: +result[3],
        voltage: +result[4],
        current_mA: +result[5],
        total_Whr: +result[6],
        temperature: +result[7]
    }
}

export const statusTotalTransmit = async () => {
    const message= await exec([
        "configure",
        `show interface ethernet all`]
    )

    return parseTotalTransmit(message)
}
