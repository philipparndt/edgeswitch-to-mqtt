package edgeswitch

import (
    "bufio"
    "fmt"
    "regexp"
    "strconv"
    "strings"
)

type DeviceInfo struct {
    Intf              string
    Detection         string
    Class             string
    ConsumedW         float64
    VoltageV          float64
    CurrentmA         float64
    ConsumedMeterWhr  float64
    TemperatureC      float64
}

// ParseDeviceInfo
// Parse the output of the "show poe interface" command
// and return a slice of DeviceInfo structs
//
// Data example:
//
// Intf      Detection      Class   Consumed(W) Voltage(V) Current(mA) Consumed Meter(Whr) Temperature(C)
// --------- -------------- ------- ----------- ---------- ----------- ------------------ --------------
// 0/10      Good           Class0         1.83      53.46       34.30              47.31             39
func ParseDeviceInfo(data string) ([]DeviceInfo, error) {
    var devices []DeviceInfo

    scanner := bufio.NewScanner(strings.NewReader(data))
    re := regexp.MustCompile(`^\d+/\d+.*`)

    for scanner.Scan() {
        line := scanner.Text()

        if re.MatchString(line) {
            deviceInfo, err := parseLine(line)
            if err != nil {
                fmt.Println("Error parsing line:", err, line)
                continue
            }
            devices = append(devices, deviceInfo)
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading file:", err)
        return nil, err
    }

    return devices, nil
}

func parseLine(line string) (DeviceInfo, error) {
    r := regexp.MustCompile("\\s\\s+")
    parts := r.Split(line, -1)

    if len(parts) != 8 {
        return DeviceInfo{}, fmt.Errorf("error parsing line: %s", line)
    }

    var deviceInfo DeviceInfo
    var err error

    deviceInfo.Intf = parts[0]
    deviceInfo.Detection = parts[1]
    deviceInfo.Class = parts[2]
    deviceInfo.ConsumedW, err = strconv.ParseFloat(parts[3], 64)
    if err != nil {
      return DeviceInfo{}, err
    }
    deviceInfo.VoltageV, err = strconv.ParseFloat(parts[4], 64)
    if err != nil {
        return DeviceInfo{}, err
    }
    deviceInfo.CurrentmA, err = strconv.ParseFloat(parts[5], 64)
    if err != nil {
        return DeviceInfo{}, err
    }
    deviceInfo.ConsumedMeterWhr, err = strconv.ParseFloat(parts[6], 64)
    if err != nil {
        return DeviceInfo{}, err
    }
    deviceInfo.TemperatureC, err = strconv.ParseFloat(parts[7], 64)
    if err != nil {
        return DeviceInfo{}, err
    }

    return deviceInfo, nil
}
