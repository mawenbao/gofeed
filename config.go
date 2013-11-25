package main

import(
    "io/ioutil"
    "log"
    "encoding/json"
    "path/filepath"
)

func ParseJsonConfig(path string) (conf Config, err error) {
    if configData, err := ioutil.ReadFile(path); nil != err {
        log.Printf("%s: %s", path, err)
    } else {
        err = json.Unmarshal(configData, &conf)
        if nil != err {
            log.Printf("%s: %s", path, err)
        }
    }

    // default settings
    // convert CacheDB to absolute path
    conf.CacheDB, err = filepath.Abs(conf.CacheDB)
    if nil != err {
        log.Printf("failed to abs CacheDB %s", conf.CacheDB)
        return
    }

    return
}

