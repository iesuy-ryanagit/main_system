package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"

    _ "github.com/go-sql-driver/mysql"
    "github.com/prometheus/client_golang/prometheus"
)

type Spending struct {
    ID       int    `json:"id,omitempty"`  // 追加、POST時は省略可能
    Date     string `json:"date"`
    Location string `json:"location"`
    Item     string `json:"item"`
    Amount   int    `json:"amount"`
}

var db *sql.DB

var (
    requestCount = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of /ping requests",
        },
    )
)



func handleSpending(w http.ResponseWriter, r *http.Request) {
    fmt.Println("handleSpending called:", r.Method, r.URL.Path)

    path := strings.TrimPrefix(r.URL.Path, "/spending/")
    fmt.Println("Trimmed path:", path)

    // /spending/ の POST・GET 処理
    if path == "" {
        switch r.Method {
        case "POST":
            var s Spending
            if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
                http.Error(w, "JSONエラー", http.StatusBadRequest)
                return
            }

            res, err := db.Exec("INSERT INTO spending (date, location, item, amount) VALUES (?, ?, ?, ?)",
                s.Date, s.Location, s.Item, s.Amount)
            if err != nil {
                http.Error(w, "DB登録失敗", http.StatusInternalServerError)
                return
            }

            lastInsertID, err := res.LastInsertId()
            if err != nil {
                http.Error(w, "ID取得失敗", http.StatusInternalServerError)
                return
            }

            s.ID = int(lastInsertID)

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusCreated)
            json.NewEncoder(w).Encode(s)

        case "GET":
            rows, err := db.Query("SELECT id, date, location, item, amount FROM spending ORDER BY date DESC")
            if err != nil {
                http.Error(w, "DB取得失敗", http.StatusInternalServerError)
                return
            }
            defer rows.Close()

            var result []Spending
            for rows.Next() {
                var s Spending
                if err := rows.Scan(&s.ID, &s.Date, &s.Location, &s.Item, &s.Amount); err != nil {
                    http.Error(w, "DB読み込みエラー", http.StatusInternalServerError)
                    return
                }
                result = append(result, s)
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(result)

        default:
            http.Error(w, "未対応メソッド", http.StatusMethodNotAllowed)
        }
        return
    }

    if r.Method == "PUT" {
        id, err := strconv.Atoi(path)
        if err != nil {
            http.Error(w, "IDが不正です", http.StatusBadRequest)
            return
        }
    
        var s Spending
        if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
            http.Error(w, "JSONエラー", http.StatusBadRequest)
            return
        }
    
        _, err = db.Exec("UPDATE spending SET date = ?, location = ?, item = ?, amount = ? WHERE id = ?",
            s.Date, s.Location, s.Item, s.Amount, id)
        if err != nil {
            http.Error(w, "更新失敗", http.StatusInternalServerError)
            return
        }
    
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "更新完了 (id=%d)", id)
        return
    }

    // /spending/{id} の DELETE 処理
    if r.Method == "DELETE" {
        id, err := strconv.Atoi(path)
        if err != nil {
            http.Error(w, "IDが不正です", http.StatusBadRequest)
            return
        }

        res, err := db.Exec("DELETE FROM spending WHERE id = ?", id)
        if err != nil {
            http.Error(w, "削除失敗", http.StatusInternalServerError)
            return
        }

        rowsAffected, err := res.RowsAffected()
        if err != nil {
            http.Error(w, "削除処理中にエラー", http.StatusInternalServerError)
            return
        }

        if rowsAffected == 0 {
            http.Error(w, "指定されたIDは存在しません", http.StatusNotFound)
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "削除完了 (id=%d)", id)
        return
    }

    http.Error(w, "不正なパスまたは未対応メソッド", http.StatusNotFound)
}

func handleMonthlySummary(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query(`
        SELECT DATE_FORMAT(date, '%Y-%m') AS month, SUM(amount)
        FROM spending
        GROUP BY month
        ORDER BY month DESC
    `)
    if err != nil {
        http.Error(w, "集計失敗", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    type Summary struct {
        Month string `json:"month"`
        Total int    `json:"total"`
    }

    var summaries []Summary
    for rows.Next() {
        var s Summary
        if err := rows.Scan(&s.Month, &s.Total); err != nil {
            http.Error(w, "集計読み込み失敗", http.StatusInternalServerError)
            return
        }
        summaries = append(summaries, s)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(summaries)
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("withCORS called for", r.Method, r.URL.Path)
        w.Header().Set("Access-Control-Allow-Origin", "*") // どこからでもアクセスOK
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS") // DELETEを追加
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        // ブラウザの事前OPTIONSリクエストに対応
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        h(w, r)
    }
}

func main() {
    podName := os.Getenv("POD_NAME")
    dsn := os.Getenv("MYSQL_DSN")

    var err error
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("DB接続失敗:", err)
    }

    if err = db.Ping(); err != nil {
        log.Fatal("DBにPing失敗:", err)
    }

    createTable := `
    CREATE TABLE IF NOT EXISTS spending (
        id INT AUTO_INCREMENT PRIMARY KEY,
        date DATE NOT NULL,
        location VARCHAR(255),
        item VARCHAR(255),
        amount INT
    );
    `
    if _, err := db.Exec(createTable); err != nil {
        log.Fatal("テーブル作成失敗:", err)
    }

    createTable := `
    CREATE TABLE IF NOT EXISTS  (
        id INT AUTO_INCREMENT PRIMARY KEY,
        date DATE NOT NULL,
        location VARCHAR(255),
        item VARCHAR(255),
        amount INT
    );
    `
    if _, err := db.Exec(createTable); err != nil {
        log.Fatal("テーブル作成失敗:", err)
    }

    defer db.Close()

    prometheus.MustRegister(requestCount)

    http.HandleFunc("/spending/", withCORS(handleSpending))
    http.HandleFunc("/summary/monthly", withCORS(handleMonthlySummary))
    http.HandleFunc("/ping", withCORS(func(w http.ResponseWriter, r *http.Request) {
        requestCount.Inc()
        fmt.Fprintf(w, "pong from %s", podName)
    }))

    fmt.Println("Server running on :8080 by v2")
    http.ListenAndServe(":8080", nil)
}
