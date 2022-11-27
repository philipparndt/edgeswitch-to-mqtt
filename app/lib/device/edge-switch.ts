import { NodeSSH } from "node-ssh"
import { getAppConfig } from "../config/config"

const exec = async (commands: string[], config = getAppConfig().edgeswitch): Promise<string> => {
    const ssh = new NodeSSH()
    const messages: string[] = []

    const client = await ssh.connect({
        host: config.ip,
        username: config.username,
        password: config.password
    })

    return new Promise((resolve, reject) => {
        client.requestShell().then(async (shell) => {
            shell.on("data", (data: Buffer) => {
                messages.push(data.toString("utf8"))
            })
            shell.on("close", async () => {
                resolve(messages.join(""))
            })

            for (const command of commands) {
                await shell.write(`${command}\n`)
            }
            await shell.end()
        })
    })
}

export const turnOff = async (port: number) => {
    await exec([
        "configure",
        `interface 0/${port}`,
        "poe opmode shutdown"]
    )
}

export const turnOn = async (port: number) => {
    await exec([
        "configure",
        `interface 0/${port}`,
        "poe opmode auto"]
    )
}

export const status = async (port: number) => {
    const message = await exec([
        "configure",
        `show poe status 0/${port}`]
    )

    const result = message
        .split("\n")
        .filter((line) => line.startsWith(`0/${port}`))
        .map(line => line.trim())
        .map((line) => line.split(/\s+/g))[0]

    if (!result || result.length === 0) {
        return undefined
    }

    return {
        interface: result[0],
        detection: result[1],
        status: result[1].toLowerCase() === "good",
        class: result[2],
        energy: +result[3],
        voltage: +result[4],
        current_mA: +result[5],
        total_Whr: +result[6],
        temperature: +result[7]
    }
}
