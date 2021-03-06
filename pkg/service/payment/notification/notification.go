package notification

import (
	"errors"
	"io"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/service"
	notificationV2 "github.com/fritzpay/paymentd/pkg/service/payment/notification/v2"
)

var (
	ErrInvalidNotificationVersion = errors.New("invalid notification version")
)

type NewNotificationFunc func(encPaymentID payment.PaymentID, p *payment.Payment) (Notification, error)

type Notification interface {
	service.Signable
	SetTransactions(payment.PaymentTransactionList)
	Sign(time.Time, string, []byte) error
	Reader() io.ReadCloser
	Identification() string
}

func NotificationByVersion(ver string) (NewNotificationFunc, error) {
	switch ver {
	case "2":
		return NewNotificationFunc(func(encPaymentID payment.PaymentID, p *payment.Payment) (Notification, error) {
			return notificationV2.New(encPaymentID, p)
		}), nil
	default:
		return nil, ErrInvalidNotificationVersion
	}
}
