
export type TransmitData = {
    port: string
    bytesTx: string
    bytesRx: string
    packetsTx: string
    packetsRx: string
}

export const parseTotalTransmit = (message: string) => {
    const result = (message ?? "")
        .split("\n")
        .filter((line) => line.match(/^\d+\/\d+.*/))
        .map(line => line.trim())
        .map((line) => line.split(/\s+/g))

    const resultData: any = {}

    for (const data of result) {
        resultData[data[0]] = {
            port: data[0],
            bytesTx: data[1],
            bytesRx: data[2],
            packetsTx: data[3],
            packetsRx: data[4]
        }
    }

    return resultData
}
