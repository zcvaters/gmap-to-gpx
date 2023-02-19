package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slog"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"
)

func CreateNewLogger(productionLogger bool) *zap.SugaredLogger {
	zapLog, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to configure logger %v", err)
	}

	if productionLogger {
		zapLog = zapLog.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
			if entry.Level > 0 {
				SendWebhookMsg(entry)
			}
			return nil
		}))

	}

	if !productionLogger {
		zapLog, err = zap.NewDevelopment()
		if err != nil {
			log.Fatalf("failed to configure development logger %v", err)
		}
	}

	defer func(zapLog *zap.Logger) {
		err := zapLog.Sync()
		if err != nil && !errors.Is(err, syscall.ENOTTY) {
			log.Printf("failed to sync zap log: %s\n", err)
		}
	}(zapLog)

	return zapLog.Sugar()
}

type DiscordWebhook struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url"`
	Content   string         `json:"content"`
	Embeds    []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Author struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		IconURL string `json:"icon_url"`
	} `json:"author"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Fields      []struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Inline bool   `json:"inline,omitempty"`
	} `json:"fields"`
	Thumbnail struct {
		URL string `json:"url"`
	} `json:"thumbnail"`
	Image struct {
		URL string `json:"url"`
	} `json:"image"`
	Footer struct {
		Text    string `json:"text"`
		IconURL string `json:"icon_url"`
	} `json:"footer"`
}

const (
	Red     = 15548997
	DarkRed = 10038562
	Orange  = 10038562
)

func SendWebhookMsg(entry zapcore.Entry) {
	level := "Error"
	color := Red
	switch entry.Level {
	case zapcore.FatalLevel:
		level = "Fatal"
		color = DarkRed
	case zapcore.DPanicLevel:
		level = "Panic"
		color = Orange
	}
	reqBody := DiscordWebhook{
		Username: "gMapToGPX",
		Embeds: []discordEmbed{
			{
				Title:       level,
				Description: entry.Message,
				Color:       color,
				Footer: struct {
					Text    string `json:"text"`
					IconURL string `json:"icon_url"`
				}(struct {
					Text    string
					IconURL string
				}{Text: entry.Time.Format("January 02, 2006 11:06:39 EST"), IconURL: ""}),
			},
		},
	}

	embedJson, _ := json.Marshal(reqBody)
	url := os.Getenv("DISCORD_WEBHOOK")
	req, _ := http.NewRequest("POST", url, bytes.NewReader(embedJson))
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
}

func NewStructuredLogger(handler slog.Handler) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{Logger: handler})
}

type StructuredLogger struct {
	Logger slog.Handler
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	var logFields []slog.Attr
	logFields = append(logFields, slog.String("ts", time.Now().UTC().Format(time.RFC1123)))

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, slog.String("req_id", reqID))
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	handler := l.Logger.WithAttrs(append(logFields,
		slog.String("http_scheme", scheme),
		slog.String("http_proto", r.Proto),
		slog.String("http_method", r.Method),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI))))

	entry := StructuredLoggerEntry{Logger: slog.New(handler)}

	entry.Logger.LogAttrs(slog.LevelInfo, "request started")
	return &entry
}

type StructuredLoggerEntry struct {
	Logger *slog.Logger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.Logger.LogAttrs(slog.LevelInfo, "request complete",
		slog.Int("resp_status", status),
		slog.Int("resp_byte_length", bytes),
		slog.Float64("resp_elapsed_ms", float64(elapsed.Nanoseconds())/1000000.0),
	)
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger.LogAttrs(slog.LevelInfo, "",
		slog.String("stack", string(stack)),
		slog.String("panic", fmt.Sprintf("%+v", v)),
	)
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.
//
// This is a useful pattern to use to set state on the entry as it
// passes through the handler chain, which at any point can be logged
// with a call to .Print(), .Info(), etc.

func GetLogEntry(r *http.Request) *slog.Logger {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	return entry.Logger
}

func LogEntrySetField(r *http.Request, key string, value interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.With(key, value)
	}
}

func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		for k, v := range fields {
			entry.Logger = entry.Logger.With(k, v)
		}
	}
}
