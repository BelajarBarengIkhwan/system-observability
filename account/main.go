package main

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/ihsansolusi/erd/devday/non-functional-test/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer        trace.Tracer
	propagator    propagation.TextMapPropagator
	statusOptions = []int{http.StatusOK, http.StatusBadRequest}
)

func main() {
	ctx := context.Background()
	propagator = telemetry.NewTelemetryPropagators()
	tp := telemetry.NewHTTPTelemetryProvider("localhost:4318", "account-service", ctx)
	tracer = tp.Tracer("main")

	api := fiber.New()
	api.Use(otelfiber.Middleware(
		otelfiber.WithTracerProvider(tp),
		otelfiber.WithPropagators(propagator),
	))
	api.Get("/validate/:acc", validasi)
	api.Listen(":3001")
}

func validasi(c *fiber.Ctx) (err error) {
	accountNo := c.Params("acc")
	validasiRekening(accountNo, c.UserContext())
	return c.SendStatus(statusOptions[rand.Intn(2)])
}

func validasiRekening(acc string, ctx context.Context) {
	ctx, span := tracer.Start(ctx, "validasi-rekening")
	span.SetAttributes(
		attribute.String("account_no", acc),
	)
	defer span.End()
	longProcess(ctx)
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
