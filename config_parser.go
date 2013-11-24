package main

import(
    "io/ioutil"
    "log"
    "encoding/json"
)

func ParseJsonConfig(path string) (targets TargetSlice, err error) {
    if configData, err := ioutil.ReadFile(path); nil != err {
        log.Printf("%s: %s", path, err)
    } else {
        err = json.Unmarshal(configData, &targets)
        if nil != err {
            log.Printf("%s: %s", path, err)
        }
    }
    return
}

