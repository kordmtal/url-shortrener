package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipResponseWriter оборачивает gin.ResponseWriter для сжатия данных
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer  *gzip.Writer
	wasUsed bool // флаг: использовался ли gzip writer
}

// Write перехватывает запись данных и сжимает их, если нужно
func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	// 1. Проверяем Content-Type (чтобы не сжимать картинки или уже сжатые файлы)
	contentType := g.Header().Get("Content-Type")

	shouldCompress := strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "application/javascript") ||
		strings.Contains(contentType, "text/css") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "text/xml")

	if !shouldCompress {
		// Если сжимать не надо — пишем в оригинальный ResponseWriter как есть
		return g.ResponseWriter.Write(data)
	}

	// 2. Если сжимаем — настраиваем заголовки (только при первой записи)
	if !g.wasUsed {
		g.wasUsed = true
		g.Header().Del("Content-Length") // Старый размер уже неверный
		g.Header().Set("Content-Encoding", "gzip")
	}

	// 3. Пишем данные через gzip.Writer
	return g.writer.Write(data)
}

// WriteString — оптимизация для записи строк (тоже сжимаем)
func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

// GzipMiddleware обрабатывает сжатие запросов и ответов
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- ЧАСТЬ 1: РАСПАКОВКА ЗАПРОСА (Decompress Request) ---
		// Если клиент прислал сжатые данные (Content-Encoding: gzip)
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			// Создаем читалку gzip поверх тела запроса
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			// Подменяем тело запроса на распакованное
			// Теперь хендлер будет читать обычный JSON, даже не зная, что он был сжат
			c.Request.Body = io.NopCloser(gzipReader)

			// Не забываем закрыть reader после обработки запроса
			defer gzipReader.Close()
		}

		// --- ЧАСТЬ 2: СЖАТИЕ ОТВЕТА (Compress Response) ---
		// Если клиент НЕ поддерживает gzip, ничего не делаем
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Создаем gzip.Writer, который будет писать в оригинальный c.Writer
		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Оборачиваем ResponseWriter в нашу структуру
		gzWriter := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}
		c.Writer = gzWriter

		// Важный заголовок для прокси-серверов
		c.Header("Vary", "Accept-Encoding")

		// Передаем управление дальше (твоему хендлеру)
		c.Next()

		// Закрываем gzip только если он использовался (чтобы дописать "хвост" архива)
		if gzWriter.wasUsed {
			gz.Close()
		}
	}
}
