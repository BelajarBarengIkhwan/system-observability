package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/ihsansolusi/erd/devday/non-functional-test/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	client     resty.Client
)

type tarikTunai struct {
	AccountNo string `json:"account_no"`
	Nominal   int    `json:"nominal"`
}

func main() {
	ctx := context.Background()
	propagator = telemetry.NewTelemetryPropagators()
	tp := telemetry.NewHTTPTelemetryProvider("localhost:4318", "transaction-service", ctx)
	tracer = tp.Tracer("main")

	client = *resty.New()
	api := fiber.New()
	api.Use(otelfiber.Middleware(
		otelfiber.WithTracerProvider(tp),
		otelfiber.WithPropagators(propagator),
	))
	api.Get("/saldo/:acc", cekSaldo)
	api.Post("/tarik", tarik)
	api.Listen(":3000")
}

func cekSaldo(c *fiber.Ctx) (err error) {
	accountNo := c.Params("acc")
	cekDana(accountNo, c.UserContext())
	return c.JSON(map[string]int{"saldo": 1000})
}

func tarik(c *fiber.Ctx) (err error) {
	var payload tarikTunai
	err = c.BodyParser(&payload)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return c.JSON(map[string]string{"remark": err.Error()})
	}
	tarikDana(payload, c.UserContext())
	return c.SendStatus(http.StatusOK)
}

func cekDana(acc string, ctx context.Context) {
	ctx, span := tracer.Start(ctx, "cek-saldo")
	span.SetAttributes(
		attribute.String("account_no", acc),
	)
	defer span.End()
	shortProcess(ctx)
}

func tarikDana(tarik tarikTunai, ctx context.Context) {
	ctx, span := tracer.Start(ctx, "tarik-dana")
	defer span.End()
	span.SetAttributes(
		attribute.String("account_no", tarik.AccountNo),
		attribute.Int("nominal", tarik.Nominal),
	)
	validasiRekening(tarik.AccountNo, ctx)
	longProcess(ctx)
	shortProcess(ctx)
}

func validasiRekening(acc string, ctx context.Context) (err error) {
	ctx, span := tracer.Start(ctx, "validasi-rekening")
	defer span.End()
	request := client.R()
	propagator.Inject(ctx, propagation.HeaderCarrier(request.Header))
	_, err = request.Get(fmt.Sprintf("http://localhost:3001/validate/%s", acc))
	return
}

func randomDuration(min, max int) (duration int) {
	duration = rand.Intn(max-min) + min
	return
}

func longProcess(ctx context.Context) {
	_, span := tracer.Start(ctx, "longprocess")
	defer span.End()
	duration := randomDuration(500, 1000)
	time.Sleep(time.Duration(duration) * time.Millisecond)
}

func shortProcess(ctx context.Context) {
	_, span := tracer.Start(ctx, "shortprocess")
	defer span.End()
	duration := randomDuration(0, 500)
	time.Sleep(time.Duration(duration) * time.Millisecond)
}
