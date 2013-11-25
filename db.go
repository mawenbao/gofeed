package main

import(
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "os"
    "errors"
    "fmt"
)

// create the cache db with sqlite3 driver
func CreateDbScheme(dbPath string) (err error) {
    _, err = os.Stat(dbPath)
    if nil == err {
        // db file already exists
        log.Printf("db %s already exists", dbPath)
        return errors.New("db " + dbPath + " already exists")
    }

    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to create db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    sqlCreateTables := fmt.Sprintf(`
    CREATE TABLE %s (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        url text NOT NULL UNIQUE,
        lastmod TEXT NOT NULL,
        html BLOB NOT NULL
    );
    `, DB_HTML_CACHE_TABLE)

    _, err = db.Exec(sqlCreateTables)
    if nil != err {
        log.Printf("failed to create tables in db %s, sql is %s, error is %s", dbPath, sqlCreateTables, err)
        return
    }

    return
}

func PutHtmlCache(dbPath string, cache []HtmlCache) (err error) {
    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to open db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    sqlInsertHtml := fmt.Sprintf(`
    INSERT INTO %s (url, lastmod, html) VALUES (?, ?, ?);
    `, DB_HTML_CACHE_TABLE)

    trans, err := db.Begin()
    if nil != err {
        log.Printf("failed to start a new transaction for db %s: %s", dbPath, err)
        return
    }

    statmt, err := trans.Prepare(sqlInsertHtml)
    if nil != err {
        log.Printf("failed to prepare a new statement for db %s, sql %s: %s", dbPath, sqlInsertHtml, err)
        return
    }
    defer statmt.Close()

    for _, c := range cache {
        _, err = statmt.Exec(c.URL, c.LastModify, c.Html)
        if nil != err {
            log.Printf("failed to insert html cache: %s", err)
            return
        }
    }

    err = trans.Commit()
    if nil != err {
        log.Printf("failed to commit transaction: %s", err)
        return
    }

    return
}

func GetHtmlCacheByURL(dbPath, url string) (cache HtmlCache, err error) {
    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to open db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    sqlSelect := fmt.Sprintf("SELECT lastmod, html FROM %s WHERE url = ?", DB_HTML_CACHE_TABLE)

    statmt, err := db.Prepare(sqlSelect)
    if nil != err {
        log.Printf("failed to prepare statment %s for db %s: %s", sqlSelect, dbPath, err)
        return
    }
    defer statmt.Close()

    cache.URL = url
    err = statmt.QueryRow(url).Scan(&cache.LastModify, &cache.Html)
    if nil != err {
        log.Printf("failed to select html cache with url %s: %s", url, err)
        return
    }

    return
}

