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
	"go.opentelemetry.io/otel/codes"
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
	err = cekDana(accountNo, c.UserContext())
	if err != nil {
		return c.SendStatus(http.StatusBadRequest)
	}
	return c.JSON(map[string]int{"saldo": 1000})
}

func tarik(c *fiber.Ctx) (err error) {
	var payload tarikTunai
	err = c.BodyParser(&payload)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return c.JSON(map[string]string{"remark": err.Error()})
	}
	err = tarikDana(payload, c.UserContext())
	if err != nil {
		return c.SendStatus(http.StatusBadRequest)
	}
	return c.SendStatus(http.StatusOK)
}

func cekDana(acc string, ctx context.Context) (err error) {
	ctx, span := tracer.Start(ctx, "cek-saldo")
	span.SetAttributes(
		attribute.String("account_no", acc),
	)
	defer span.End()
	err = validasiRekening(acc, ctx)
	if err != nil {
		span.SetStatus(codes.Error, "cek dana rekening gagal")
		return
	}
	shortProcess(ctx)
	return
}

func tarikDana(tarik tarikTunai, ctx context.Context) (err error) {
	ctx, span := tracer.Start(ctx, "tarik-dana")
	defer span.End()
	span.SetAttributes(
		attribute.String("account_no", tarik.AccountNo),
		attribute.Int("nominal", tarik.Nominal),
	)
	err = validasiRekening(tarik.AccountNo, ctx)
	if err != nil {
		span.SetStatus(codes.Error, "tarik dana rekening gagal")
		return
	}
	longProcess(ctx)
	shortProcess(ctx)
	return
}

func validasiRekening(acc string, ctx context.Context) (err error) {
	ctx, span := tracer.Start(ctx, "validasi-rekening")
	defer span.End()
	request := client.R()
	propagator.Inject(ctx, propagation.HeaderCarrier(request.Header))
	resp, err := request.Get(fmt.Sprintf("http://localhost:3001/validate/%s", acc))
	if resp.StatusCode() != 200 {
		span.SetStatus(codes.Error, "validasi rekening gagal")
		return fmt.Errorf("validasi rekening gagal")
	}
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
