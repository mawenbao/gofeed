package main

import(
    "io/ioutil"
    "log"
    "encoding/json"
)

func ParseJsonConfig(path string) (targets TargetSlice) {
    if configData, err := ioutil.ReadFile(path); nil != err {
        log.Fatalf("%s: %s\n", path, err)
    } else {
        err = json.Unmarshal(configData, &targets)
        if nil != err {
            log.Fatalf("%s: %s\n", path, err)
        }
    }
    return
}

