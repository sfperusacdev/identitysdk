package asistencia

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"
	"github.com/sfperusacdev/identitysdk/telemetry"
	"github.com/user0608/goones/types"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type resumenHorariosClientStub struct {
	asistenciapb.TrabajadoresHorariosServiceClient
	err error
	t   *testing.T
}

func (s resumenHorariosClientStub) ListarTrabajadoresHorarios(ctx context.Context, req *asistenciapb.ListarTrabajadoresHorariosRequest, opts ...grpc.CallOption) (*asistenciapb.ListarTrabajadoresHorariosResponse, error) {
	if len(req.TrabajadoresCodigos) != 1 || req.TrabajadoresCodigos[0] != "empresa.trabajador-privado" {
		s.t.Fatal("la instrumentacion modifico los trabajadores solicitados")
	}
	found := false
	for _, opt := range opts {
		if size, ok := opt.(grpc.MaxRecvMsgSizeCallOption); ok && size.MaxRecvMsgSize == 16*1024*1024 {
			found = true
		}
	}
	if !found {
		s.t.Fatal("falta limite de recepcion de 16 MiB")
	}
	return &asistenciapb.ListarTrabajadoresHorariosResponse{Data: []*asistenciapb.ResultTrabajadoresHorarios{{
		TrabajadorCodigo: req.TrabajadoresCodigos[0],
		Fechas:           []*asistenciapb.Dia{{HorasRegularesDia: "8"}},
	}}}, s.err
}

func TestResumenHorarios_RangoJornales_Traces(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "respuesta correcta"},
		{name: "error rpc", err: status.Error(codes.ResourceExhausted, "error privado")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var spans []*tracepb.Span
			collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				request := &collectortrace.ExportTraceServiceRequest{}
				if err := proto.Unmarshal(body, request); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				mu.Lock()
				defer mu.Unlock()
				for _, resource := range request.ResourceSpans {
					for _, scope := range resource.ScopeSpans {
						spans = append(spans, scope.Spans...)
					}
				}
				w.Header().Set("Content-Type", "application/x-protobuf")
			}))
			defer collector.Close()
			config := telemetry.DefaultConfig()
			config.Enabled = true
			config.OTLP.Endpoint = collector.URL
			config.Traces.SampleRatio = 1
			config.Logs.ExportOTLP = false
			service, err := telemetry.NewTelemetryService(t.Context(), config, telemetry.ServiceInfo{Name: "test"})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = service.Shutdown(ctx)
			})
			ctx, parent := service.TracerProvider().Tracer("test").Start(t.Context(), "peticion")
			client := NewAsistenciaService(resumenHorariosClientStub{t: t, err: tt.err})
			fecha := types.NewDateOnly(time.Now())
			results, err := client.ResumenHorarios_RangoJornales(ctx, fecha, fecha, []string{"empresa.trabajador-privado"})
			if status.Code(err) != status.Code(tt.err) {
				t.Fatalf("error=%v, esperado %v", err, tt.err)
			}
			if tt.err == nil && (len(results) != 1 || len(results[0].Fechas) != 1 || results[0].Fechas[0].HorasRegularesDia.String() != "8") {
				t.Fatalf("resultado alterado: %+v", results)
			}
			parent.End()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := service.Shutdown(shutdownCtx); err != nil {
				t.Fatal(err)
			}
			mu.Lock()
			defer mu.Unlock()
			byName := make(map[string]*tracepb.Span)
			parentTraceID := parent.SpanContext().TraceID()
			for _, span := range spans {
				byName[span.Name] = span
				if !bytes.Equal(span.TraceId, parentTraceID[:]) {
					t.Error("se perdio la traza padre")
				}
				if bytes.Contains([]byte(span.String()), []byte("privado")) {
					t.Error("se exportaron datos privados")
				}
			}
			root, rpc := byName["asistencia.resumen_horarios"], byName["asistencia.resumen_horarios.rpc"]
			if root == nil || rpc == nil {
				t.Fatalf("faltan spans: %v", byName)
			}
			if !bytes.Equal(rpc.ParentSpanId, root.SpanId) {
				t.Error("el span RPC no es hijo del resumen")
			}
			if tt.err != nil && (root.GetStatus().GetCode() != tracepb.Status_STATUS_CODE_ERROR || rpc.GetStatus().GetCode() != tracepb.Status_STATUS_CODE_ERROR) {
				t.Error("no se registro el estado de error")
			}
			if tt.err == nil && byName["asistencia.resumen_horarios.convertir"] == nil {
				t.Error("falta span de conversion")
			}
			attributes := make(map[string]int64)
			for _, attr := range root.Attributes {
				attributes[attr.Key] = attr.Value.GetIntValue()
			}
			if attributes["horarios.trabajadores_solicitados"] != 1 {
				t.Error("falta cantidad de trabajadores solicitados")
			}
			if tt.err == nil && (attributes["horarios.trabajadores_devueltos"] != 1 || attributes["horarios.respuesta_bytes"] <= 0) {
				t.Error("faltan cantidades de respuesta")
			}
		})
	}
}
