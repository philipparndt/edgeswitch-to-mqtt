package edgeswitch

import (
    "bufio"
    "fmt"
    "regexp"
    "rnd7/edgeswitch-mqtt/logger"
    "strings"
)

type ChannelData struct {
    Port        string
    BytesTx     int64
    BytesRx     int64
    PacketsTx   int64
    PacketsRx   int64
}

func ParseChannelData(data string) ([]ChannelData, error) {
    var channelDataList []ChannelData

    scanner := bufio.NewScanner(strings.NewReader(data))
    re := regexp.MustCompile(`^\d+/\d+.*`)

    for scanner.Scan() {
        line := scanner.Text()

        if re.MatchString(line) {
            channelData, err := parseChannelLine(line)
            if err != nil {
                logger.Error("Error parsing line:", err)
                continue
            }
            channelDataList = append(channelDataList, channelData)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return channelDataList, nil
}

func parseChannelLine(line string) (ChannelData, error) {
    var channelData ChannelData

    // Use fmt.Sscanf to parse values from the line
    _, err := fmt.Sscanf(line, "%s %d %d %d %d",
        &channelData.Port, &channelData.BytesTx, &channelData.BytesRx, &channelData.PacketsTx, &channelData.PacketsRx)

    if err != nil {
        return ChannelData{}, err
    }

    return channelData, nil
}
