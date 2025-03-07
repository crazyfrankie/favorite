package rpc

import (
	"context"
	"fmt"
	"net"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/favorite/internal/biz/service"
	"github.com/crazyfrankie/favorite/internal/config"
	"github.com/crazyfrankie/favorite/pkg/registry"
	"github.com/crazyfrankie/favorite/rpc_gen/favorite"
)

var (
	PromRegistry = prometheus.NewRegistry()
)

type Server struct {
	*grpc.Server
	Port     string
	registry *registry.ServiceRegistry
}

func NewServer(f *service.FavoriteServer, client *clientv3.Client) *Server {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	traceId := func(ctx context.Context) logging.Fields {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}

		return nil
	}

	favoriteMetrics := grpcprom.NewServerMetrics()
	PromRegistry.MustRegister(favoriteMetrics)

	labelsFromContext := func(ctx context.Context) prometheus.Labels {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	tp := initTracerProvider("service/favorite")
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			favoriteMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.UnaryServerInterceptor(interceptorLogger(logger), logging.WithFieldsFromContext(traceId)),
		),
	)
	favorite.RegisterFavoriteServiceServer(s, f)

	rgy, err := registry.NewServiceRegistry(client)
	if err != nil {
		panic(err)
	}

	return &Server{
		Server:   s,
		Port:     config.GetConf().Server.Port,
		registry: rgy,
	}
}

func (s *Server) Serve() error {
	conn, err := net.Listen("tcp", s.Port)
	if err != nil {
		return err
	}

	err = s.registry.Register()
	if err != nil {
		return err
	}

	return s.Server.Serve(conn)
}

func (s *Server) Shutdown() {
	err := s.registry.UnRegister()
	if err != nil {
		zap.L().Error("Failed to unregister", zap.Error(err))
	}

	s.Server.GracefulStop()
	s.Server.Stop()
}

func interceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			val := fields[i+1]

			switch v := val.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			case any:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch level {
		case logging.LevelDebug:
			logger.Debug(msg, f...)
		case logging.LevelInfo:
			logger.Info(msg, f...)
		case logging.LevelWarn:
			logger.Warn(msg, f...)
		case logging.LevelError:
			logger.Error(msg, f...)
		default:
			panic(fmt.Sprintf("unknown level %v", level))
		}
	})
}

func initTracerProvider(servicename string) *trace.TracerProvider {
	res, err := newResource(servicename, "v0.0.1")
	if err != nil {
		fmt.Printf("failed create resource, %s", err)
	}

	tp, err := newTraceProvider(res)
	if err != nil {
		panic(err)
	}

	return tp
}

func newResource(servicename, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String(servicename),
			semconv.ServiceVersionKey.String(serviceVersion)))
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)), trace.WithResource(res))

	return traceProvider, nil
}
