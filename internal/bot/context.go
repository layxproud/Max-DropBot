package bot

import (
	"context"
	"time"
)

// noDeadlineCtx оборачивает контекст, скрывая его дедлайн.
// Используется для long-polling: библиотека сама управляет таймаутом запроса,
// а отмена (Done/Err) родительского контекста по-прежнему работает.
type noDeadlineCtx struct{ context.Context }

func (noDeadlineCtx) Deadline() (time.Time, bool) { return time.Time{}, false }

func withoutDeadline(ctx context.Context) context.Context {
	return noDeadlineCtx{ctx}
}
