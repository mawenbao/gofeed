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
        url TEXT NOT NULL UNIQUE,
        title TEXT,
        cache_control TEXT,
        lastmod TEXT,
        etag TEXT,
        expires TEXT,
        html BLOB
    );
    `, DB_HTML_CACHE_TABLE)

    _, err = db.Exec(sqlCreateTables)
    if nil != err {
        log.Printf("failed to create tables in db %s, sql is %s, error is %s", dbPath, sqlCreateTables, err)
        return
    }

    return
}

func ExecQuerySQL(dbPath string, expectSize int, sqlStr string, args ...interface{}) (caches []HtmlCache, err error) {
    _, err = os.Stat(dbPath)
    if nil != err {
        // db file not exists
        log.Printf("db %s not exists", dbPath)
        return
    }

    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to open db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    statmt, err := db.Prepare(sqlStr)
    if nil != err {
        log.Printf("failed to prepare statment %s for db %s: %s", sqlStr, dbPath, err)
        return
    }
    defer statmt.Close()

    rows, err := statmt.Query(args...)
    if nil != err {
        log.Printf("failed to query with statment %s, %s", sqlStr, err)
        return
    }
    defer rows.Close()

    if expectSize > 0 {
        caches = make([]HtmlCache, expectSize)
    }
    rowInd := 0
    for rows.Next() {
        var c HtmlCache
        err = rows.Scan(
            &c.URL,
            &c.CacheControl,
            &c.LastModified,
            &c.Etag,
            &c.Expires,
            &c.Html)
        if nil != err {
            log.Printf("failed to scan data from result row: %s", err)
            return
        }
        caches = append(caches[:rowInd], c)
        rowInd += 1
    }

    // no result is also an error
    if 0 == rowInd {
        err = DBNoRecordError {}
        return
    }

    return
}

func ExecInsertUpdateSQL(caches []HtmlCache, dbPath string, sqlStr string) (err error) {
    _, err = os.Stat(dbPath)
    if nil != err {
        // db file not exists
        log.Printf("db %s not exists", dbPath)
        return
    }

    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to open db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    trans, err := db.Begin()
    if nil != err {
        log.Printf("failed to start a new transaction for db %s: %s", dbPath, err)
        return
    }

    statmt, err := trans.Prepare(sqlStr)
    if nil != err {
        log.Printf("failed to prepare a new statement for db %s, sql %s: %s", dbPath, sqlStr, err)
        return
    }
    defer statmt.Close()

    urls := ""
    for _, c := range caches {
        _, err = statmt.Exec(c.URL, c.Title, c.CacheControl, c.LastModified, c.Etag, c.Expires, c.Html)
        urls += c.URL + " "
        if nil != err {
            log.Printf("failed to exec insert/update sql %s: %s", sqlStr, err)
            return
        }
    }

    err = trans.Commit()
    if nil != err {
        log.Printf("failed to save urls %s: %s", urls, err)
        return
    }

    return
}

func GetHtmlCacheByURL(dbPath, url string) (cache HtmlCache, err error) {
    htmlCacheSlice, err := ExecQuerySQL(
        dbPath,
        1,
        fmt.Sprintf("SELECT url, cache_control, lastmod, etag, expires, html FROM %s WHERE url = ?", DB_HTML_CACHE_TABLE),
        url)

    if nil != err {
        switch err.(type) {
        case DBNoRecordError:
            if *gVerbose {
                log.Printf("cache not found for %s", url)
            }
        default:
            log.Printf("failed to get cache from db %s by url %s: %s", dbPath, url, err)
        }
    }

    return htmlCacheSlice[0], err
}

func PutHtmlCache(dbPath string, caches []HtmlCache) (err error) {
    sqlInsertHtml := fmt.Sprintf(`
    INSERT INTO %s (url, title, cache_control, lastmod, etag, expires, html) VALUES (?, ?, ?, ?, ?, ?, ?);
    `, DB_HTML_CACHE_TABLE)

    err = ExecInsertUpdateSQL(caches, dbPath, sqlInsertHtml)
    if nil != err {
        log.Printf("failed to insert cache records to db %s: %s", dbPath, err)
        return
    }

    if *gVerbose {
        for _, c := range caches {
            log.Printf("successully saved cache of %s", c.URL)
        }
    }
    return
}

func UpdateHtmlCache(dbPath string, caches []HtmlCache) (err error) {
    sqlUpdateHtml := ""
    for _, c := range caches {
        sqlUpdateHtml += fmt.Sprintf(`
        UPDATE %s SET url = ?, title = ?, cache_control = ?, lastmod = ?, etag = ?, expires = ?, html = ? WHERE url = '%s';
        `, DB_HTML_CACHE_TABLE, c.URL)
    }
    err = ExecInsertUpdateSQL(caches, dbPath, sqlUpdateHtml)
    if nil != err {
        log.Printf("failed to update cache records to db %s: %s", dbPath, err)
        return
    }

    if *gVerbose {
        for _, c := range caches {
            log.Printf("successully updated cache of %s", c.URL)
        }
    }
    return
}

