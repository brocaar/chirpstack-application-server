---
title: Frame logging
menu:
    main:
        parent: use
        weight: 11
description: Show live LoRaWAN frames for debugging device and gateway behavior.
---

# Frame logging

ChirpStack Application Server makes it possible to log frames sent and received by a gateway
or device in realtime. To use this feature, you first need to go to the gateway
or device detail page. Once you are on this page, open the **LoRaWAN frames**
tab.

**Note:** This is a feature for debugging the communication between the
network-server and your devices. Do not use this for integration with your
applications.

As soon as you open this page, ChirpStack Application Server will subscribe to the frames
sent / received by the selected gateway or device. Once a frame is sent or
received, it will be displayed without the need to refresh the page.

### Gateway frame logs

The frame logs view on the gateway detail page will display all frames sent
and received by the selected gateway.

### Device frame logs

The frame logs view on the device detail page will display only the frames
that could be related to a device. This means that in case you want to debug
MIC issues, this view might not show this information and it is better to use
the gateway view.

## Exposed information

Note that all the displayed data can be expanded by clicking on each key.
E.g. **> phyPayload: {} 3 keys** means you can expand this **phyPayload**
item as it has three sub-items. This makes it possible to inspect
(unless encrypted):

* TX related meta-data (frequency, data-rate, ...)
* RX related-meta-data (RSSI, SNR, timestamps, ...)
* LoRaWAN PHYPayload frames and mac-commands (unless encrypted)

## Storing frames

Frames displayed are not stored in a database. As soon as you refresh the page,
all data is gone.
If you need to store these frames, you can use the **download** button which
will generate a JSON formatted file containing all the frames that are
logged since opening the **Live frame logs** tab.

### Example JSON output

{{<highlight json>}}
[
    {
        "uplinkMetaData": {
            "rxInfo": [
                {
                    "mac": "0102030405060708",
                    "time": "2018-02-13T14:00:00.683952Z",
                    "timeSinceGPSEpoch": "",
                    "timestamp": 123456,
                    "rssi": -150,
                    "loRaSNR": -7,
                    "board": 0,
                    "antenna": 0
                }
            ],
            "txInfo": {
                "frequency": 868300000,
                "dataRate": {
                    "modulation": "LORA",
                    "bandwidth": 125,
                    "spreadFactor": 12,
                    "bitrate": 0
                },
                "codeRate": "4/5"
            }
        },
        "phyPayload": {
            "mhdr": {
                "mType": "UnconfirmedDataUp",
                "major": "LoRaWANR1"
            },
            "macPayload": {
                "fhdr": {
                    "devAddr": "01020304",
                    "fCtrl": {
                        "adr": false,
                        "adrAckReq": false,
                        "ack": false,
                        "fPending": false
                    },
                    "fCnt": 1234,
                    "fOpts": null
                },
                "fPort": 1,
                "frmPayload": [
                    {
                        "bytes": "..."
                    }
                ]
            },
            "mic": "04030201"
        }
    }
]
{{< /highlight >}}
