package paypal_rest

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	tmpl "github.com/fritzpay/paymentd/pkg/template"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"gopkg.in/inconshreveable/log15.v2"
)

func (d *Driver) getTemplate(t *template.Template, tmplDir, locale, baseName string) (err error) {
	tmplFile, err := tmpl.TemplateFileName(tmplDir, locale, defaultLocale, baseName)
	if err != nil {
		return err
	}
	tmplB, err := ioutil.ReadFile(tmplFile)
	if err != nil {
		return err
	}
	tmplLocale := path.Base(path.Ext(tmplFile))
	t.Funcs(template.FuncMap(map[string]interface{}{
		"staticPath": func() (string, error) {
			url, err := d.mux.Get("staticHandler").URLPath()
			if err != nil {
				return "", err
			}
			return url.Path, nil
		},
		"locale": func() string {
			return tmplLocale
		},
	}))
	_, err = t.Parse(string(tmplB))
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) templatePaymentData(p *payment.Payment) map[string]interface{} {
	tmplData := make(map[string]interface{})
	if p != nil {
		tmplData["payment"] = p
		tmplData["paymentID"] = d.paymentService.EncodedPaymentID(p.PaymentID())
		tmplData["amount"] = p.DecimalRound(2)
	}
	tmplData["timestamp"] = time.Now().Unix()
	return tmplData
}

func writeTemplateBuf(log log15.Logger, w io.Writer, tmpl *template.Template, tmplData interface{}) error {
	buf := buffer()
	err := tmpl.Execute(buf, tmplData)
	if err != nil {
		log.Error("error on template", log15.Ctx{"err": err})
		return ErrInternal
	}
	_, err = io.Copy(w, buf)
	putBuffer(buf)
	buf = nil
	if err != nil {
		log.Error("error writing buffered output", log15.Ctx{"err": err})
	}
	return nil
}

// InitPageHandler serves the init page (loading screen)
func (d *Driver) InitPageHandler(p *payment.Payment) http.Handler {
	const baseName = "init.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "InitPageHandler"})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.New("init")
		err := d.getTemplate(tmpl, d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := d.templatePaymentData(p)
		err = writeTemplateBuf(log, w, tmpl, tmplData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// InternalErrorHandler serves the page notifying the user about a (critical)
// internal error. The payment can not continue.
//
// It can handle a nil payment parameter.
func (d *Driver) InternalErrorHandler(p *payment.Payment) http.Handler {
	const baseName = "internal_error.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "InternalErrorHandler"})

		tmplData := d.templatePaymentData(p)
		// do log so we can find the timestamp in the logs
		log.Error("internal error", log15.Ctx{"timestamp": tmplData["timestamp"]})
		w.WriteHeader(http.StatusInternalServerError)
		locale := defaultLocale
		if p != nil {
			locale = p.Config.Locale.String
		}
		tmpl := template.New("internal_error")
		err := d.getTemplate(tmpl, d.tmplDir, locale, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			return
		}
		writeTemplateBuf(log, w, tmpl, tmplData)
	})
}

func (d *Driver) PaymentErrorHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func (d *Driver) BadRequestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func (d *Driver) NotFoundHandler(p *payment.Payment) http.Handler {
	const baseName = "not_found.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "NotFoundHandler"})

		tmplData := d.templatePaymentData(p)
		// do log so we can find the timestamp in the logs
		log.Warn("payment not found", log15.Ctx{"timestamp": tmplData["timestamp"]})
		w.WriteHeader(http.StatusNotFound)
		locale := defaultLocale
		if p != nil {
			locale = p.Config.Locale.String
		}
		tmpl := template.New("not_found")
		err := d.getTemplate(tmpl, d.tmplDir, locale, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			return
		}
		writeTemplateBuf(log, w, tmpl, tmplData)
	})
}

func (d *Driver) CancelPageHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{
			"method":    "CancelPageHandler",
			"projectID": p.ProjectID(),
			"paymentID": p.PaymentID(),
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.New("cancel")
		const baseName = "cancel.html.tmpl"
		err := d.getTemplate(tmpl, d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := d.templatePaymentData(p)
		err = writeTemplateBuf(log, w, tmpl, tmplData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (d *Driver) ReturnPageHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{
			"method":    "ReturnPageHandler",
			"projectID": p.ProjectID(),
			"paymentID": p.PaymentID(),
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.New("return")
		const baseName = "return.html.tmpl"
		err := d.getTemplate(tmpl, d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := d.templatePaymentData(p)
		err = writeTemplateBuf(log, w, tmpl, tmplData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (d *Driver) SuccessHandler(p *payment.Payment) http.Handler {
	const baseName = "success.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "SuccessHandler"})

		tmplData := d.templatePaymentData(p)
		locale := defaultLocale
		if p != nil {
			locale = p.Config.Locale.String
		}
		tmpl := template.New("success")
		err := d.getTemplate(tmpl, d.tmplDir, locale, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			return
		}
		writeTemplateBuf(log, w, tmpl, tmplData)
	})
}

func (d *Driver) ApprovalHandler(tx *Transaction, p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{
			"method":               "ApprovalHandler",
			"projectID":            p.ProjectID(),
			"paymentID":            p.PaymentID(),
			"transactionTimestamp": tx.Timestamp.UnixNano(),
		})
		links, err := tx.PayPalLinks()
		if err != nil {
			log.Error("transaction links error", log15.Ctx{"err": err})
			d.PaymentErrorHandler(p).ServeHTTP(w, r)
			return
		}
		if links["approval_url"] == nil {
			log.Error("no approval URL")
			d.PaymentErrorHandler(p).ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, links["approval_url"].HRef, http.StatusTemporaryRedirect)
	})
}

func (d *Driver) PaymentStatusHandler(p *payment.Payment) http.Handler {
	switch p.Status {
	case payment.PaymentStatusCancelled:
		return d.CancelPageHandler(p)
	case payment.PaymentStatusPaid, payment.PaymentStatusAuthorized:
		return d.SuccessHandler(p)
	case payment.PaymentStatusError:
		return d.PaymentErrorHandler(p)
	default:
		d.log.Warn("unknown payment status", log15.Ctx{
			"method":                   "PaymentStatusHandler",
			"paymentTransactionStatus": p.Status,
		})
		return d.PaymentErrorHandler(p)
	}
}

// the returned handler will serve the appropriate init action based on the current
// paypal transaction status
func (d *Driver) statusHandler(tx *Transaction, p *payment.Payment, defaultHandler http.Handler) http.Handler {
	return d.pollStatusHandler(tx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch tx.Type {
		case TransactionTypeCreatePaymentResponse:
			if tx.PaypalState.String == "created" {
				d.ApprovalHandler(tx, p).ServeHTTP(w, r)
				return
			}
			d.PaymentStatusHandler(p).ServeHTTP(w, r)
		case TransactionTypeError:
			d.PaymentErrorHandler(p).ServeHTTP(w, r)
		case TransactionTypeGetPaymentResponse, TransactionTypeExecutePaymentResponse:
			d.PaymentStatusHandler(p).ServeHTTP(w, r)
		default:
			defaultHandler.ServeHTTP(w, r)
		}
	}))
}

func (d *Driver) pollStatusHandler(tx *Transaction, parent http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ajax poll?
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") &&
			!strings.Contains(r.Header.Get("Accept"), "application/json") {
			parent.ServeHTTP(w, r)
			return
		}
		// will be true when the polling (ajax) should stop and reload
		cont := true
		// wait on create payment request
		// wait on get payment request
		// wait on execute payment request
		if tx.Type == TransactionTypeCreatePayment ||
			tx.Type == TransactionTypeExecutePayment ||
			tx.Type == TransactionTypeGetPayment {
			cont = false
		}
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, "{\"c\": %t}", cont)
		if err != nil {
			d.log.Error("error writing response", log15.Ctx{
				"method": "StatusHandler",
				"err":    err,
			})
		}
		return
	})
}
