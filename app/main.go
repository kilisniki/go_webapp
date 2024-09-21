package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    // Получаем DATABASE_URL из переменных окружения
    databaseUrl := os.Getenv("DATABASE_URL")
    if databaseUrl == "" {
        log.Fatal("Переменная окружения DATABASE_URL не установлена")
    }

    // Подключаемся к базе данных
    ctx := context.Background()
    dbpool, err := pgxpool.New(ctx, databaseUrl)
    if err != nil {
        log.Fatalf("Не удалось подключиться к базе данных: %v\n", err)
    }
    defer dbpool.Close()

    // Создаем таблицу counter, если она не существует
    _, err = dbpool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS counter (
            id SERIAL PRIMARY KEY,
            count BIGINT NOT NULL
        );
    `)
    if err != nil {
        log.Fatalf("Не удалось создать таблицу: %v\n", err)
    }

    // Проверяем, существует ли запись с id = 1, если нет — создаем
    var count int64
    err = dbpool.QueryRow(ctx, "SELECT count FROM counter WHERE id = 1").Scan(&count)
    if err != nil {
        if err == pgx.ErrNoRows {
            _, err = dbpool.Exec(ctx, "INSERT INTO counter (id, count) VALUES (1, 0)")
            if err != nil {
                log.Fatalf("Не удалось вставить начальное значение счетчика: %v\n", err)
            }
            count = 0
        } else {
            log.Fatalf("Ошибка при запросе счетчика: %v\n", err)
        }
    }

    // Инициализируем Gin
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        // Инкрементируем счетчик и получаем новое значение
        err := dbpool.QueryRow(ctx, "UPDATE counter SET count = count + 1 WHERE id = 1 RETURNING count").Scan(&count)
        if err != nil {
            log.Printf("Не удалось обновить счетчик: %v\n", err)
            c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
            return
        }

        // Возвращаем "Hello World" и текущее значение счетчика
        c.String(http.StatusOK, "Hello World! Вы посетитель номер %d", count)
    })

    // Запускаем сервер
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    r.Run(":" + port)
}
