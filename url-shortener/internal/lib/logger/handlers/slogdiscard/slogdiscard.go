package slogdiscard

import(
	"log/slog"
	"context"
)
func NewDiscardLogger() *slog.Logger{
	return slog.New(NewDiscardHandler())
}

type DiscardHandler struct{}

func NewDiscardHandler()*DiscardHandler{
	return &DiscardHandler{}
}

func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error{ // Игнорируем запись журнала
	return nil
}

func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler{ // Возвращает тот же обработчик, тк нет атрибутов для сохранения
	return h
}

func (h *DiscardHandler) WithGroup(_ string) slog.Handler{ // Возвращает тот же обработчик, тк нет группы для сохранения
	return h
}

func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool{ // Возвращает false, тк запись журнала игнорируется
	return false
}