package server

import (
	"fmt"
	"html"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"

	"github.com/ingvarmattis/moving/src/transport/orders"
)

type NoopTelegramBot struct{}

func (NoopTelegramBot) NotifyNewOrder(*orders.Order) {}
func (NoopTelegramBot) Start()                       {}
func (NoopTelegramBot) Close()                       {}

// NewNoopTelegramBot returns a no-op telegram bot struct.
func NewNoopTelegramBot() *NoopTelegramBot {
	return &NoopTelegramBot{}
}

type TelegramBot struct {
	tb             *telebot.Bot
	logger         *zap.Logger
	allowedChatIDs []int64
}

func NewTelegramBot(
	logger *zap.Logger, token string, timeout time.Duration, allowedChatIDs []int64,
) (*TelegramBot, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: timeout},
	}

	tBot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot | %w", err)
	}

	tBot.Use(middleware.Whitelist(allowedChatIDs...))

	bot := &TelegramBot{
		tb:             tBot,
		logger:         logger,
		allowedChatIDs: allowedChatIDs,
	}

	tBot.Handle("/start", bot.onStart)

	return bot, nil
}

func (b *TelegramBot) onStart(ctx telebot.Context) error {
	return ctx.Send("You have subscribed to updates.")
}

func (b *TelegramBot) Start() {
	b.logger.Info("starting telegram bot")
	b.tb.Start()
}

func (b *TelegramBot) Close() {
	b.tb.Stop()
	_, _ = b.tb.Close()
}

// NotifyNewOrder sends the new order details to all allowed chats.
func (b *TelegramBot) NotifyNewOrder(order *orders.Order) {
	text := formatOrderMessage(order)
	opts := &telebot.SendOptions{ParseMode: telebot.ModeHTML}
	for _, chatID := range b.allowedChatIDs {
		recipient := &telebot.Chat{ID: chatID}
		if _, err := b.tb.Send(recipient, text, opts); err != nil {
			b.logger.Error("send new order to telegram", zap.Error(err), zap.Int64("chat_id", chatID), zap.Uint64("order_id", order.ID))
		}
	}
}

func formatOrderMessage(o *orders.Order) string {
	escape := func(s string) string { return html.EscapeString(s) }

	var b strings.Builder
	b.WriteString("<b>NEW ORDER</b> #")
	b.WriteString(fmt.Sprint(o.ID))
	b.WriteString("\n\n")
	b.WriteString("<b>From:</b> ")
	b.WriteString(escape(o.MoveFrom))
	b.WriteString("\n<b>To:</b> ")
	b.WriteString(escape(o.MoveTo))
	b.WriteString("\n")
	if !o.MoveDate.IsZero() {
		b.WriteString("<b>Date:</b> ")
		b.WriteString(escape(o.MoveDate.Format(time.DateOnly)))
		b.WriteString("\n")
	}
	if o.Phone != "" {
		b.WriteString("<b>Phone:</b> ")
		b.WriteString(escape(o.Phone))
		b.WriteString("\n")
	}
	if o.Email != nil && *o.Email != "" {
		b.WriteString("<b>Email:</b> ")
		b.WriteString(escape(*o.Email))
		b.WriteString("\n")
	}
	if o.AdditionalInfo != nil && *o.AdditionalInfo != "" {
		b.WriteString("\n<b>Description:</b>\n")
		b.WriteString(escape(*o.AdditionalInfo))
	}
	return strings.TrimSuffix(b.String(), "\n")
}
