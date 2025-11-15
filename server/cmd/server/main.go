package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	nanabushv1 "github.com/dasmlab/nanabush/server/pkg/proto/v1"
	"github.com/dasmlab/nanabush/server/pkg/service"
)

var (
	port         = flag.Int("port", 50051, "gRPC server port")
	insecureMode = flag.Bool("insecure", true, "Run server in insecure mode (no TLS)")
	
	// TLS configuration flags (for future use)
	tlsCertPath = flag.String("tls-cert", "", "Path to TLS server certificate")
	tlsKeyPath  = flag.String("tls-key", "", "Path to TLS server private key")
	tlsCAPath   = flag.String("tls-ca", "", "Path to CA certificate for client verification (mTLS)")
)

func main() {
	flag.Parse()
	
	logger := log.New(os.Stdout, "[nanabush-grpc] ", log.LstdFlags|log.Lshortfile)
	logger.Printf("Starting Nanabush gRPC server on port %d (insecure=%v)", *port, *insecureMode)
	
	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Fatalf("Failed to listen on port %d: %v", *port, err)
	}
	
	// Create gRPC server with options
	var opts []grpc.ServerOption
	
	// TODO: Configure TLS/mTLS when certificates are available
	if !*insecureMode {
		// TODO: Load TLS credentials from flags
		// For now, log warning and continue with insecure
		logger.Println("WARNING: TLS requested but not yet implemented, using insecure mode")
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	}
	
	// Create gRPC server
	s := grpc.NewServer(opts...)
	
	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	
	// Register translation service
	// TODO: Implement actual vLLM backend integration
	translationService := service.NewTranslationService(nil, logger)
	nanabushv1.RegisterTranslationServiceServer(s, translationService)
	
	// Enable reflection for grpcurl/debugging (can be disabled in production)
	reflection.Register(s)
	
	// Start periodic cleanup goroutine for expired clients
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Run cleanup every 5 minutes
		defer ticker.Stop()
		
		maxIdleTime := 15 * time.Minute // Remove clients idle for more than 15 minutes
		
		for {
			select {
			case <-ticker.C:
				translationService.CleanupExpiredClients(maxIdleTime)
			case <-cleanupCtx.Done():
				return
			}
		}
	}()
	logger.Println("Started client cleanup goroutine (runs every 5 minutes)")
	
	// Start periodic metrics logging
	metricsCtx, metricsCancel := context.WithCancel(context.Background())
	defer metricsCancel()
	
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Log metrics every minute
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Get aggregated metrics
				metrics := translationService.GetClientMetrics()
				logger.Printf("Client metrics: total_registered=%d", metrics.TotalClients)
				
				if metrics.TotalClients > 0 {
					// Log namespace distribution
					for ns, count := range metrics.ClientsByNamespace {
						logger.Printf("  By namespace: %q=%d", ns, count)
					}
					
					// Log version distribution
					for version, count := range metrics.ClientsByVersion {
						logger.Printf("  By version: %q=%d", version, count)
					}
					
					// Log heartbeat stats
					oldestAge := time.Since(metrics.OldestHeartbeat)
					newestAge := time.Since(metrics.NewestHeartbeat)
					logger.Printf("  Heartbeat stats: oldest=%v ago, newest=%v ago",
						oldestAge, newestAge)
					
					// Log individual client details (first 5 to avoid log spam)
					clients := translationService.GetRegisteredClients()
					maxLog := 5
					if len(clients) < maxLog {
						maxLog = len(clients)
					}
					for i := 0; i < maxLog; i++ {
						client := clients[i]
						lastHeartbeat := time.Since(client.LastHeartbeat)
						logger.Printf("  Client[%d]: id=%q, name=%q, last_heartbeat=%v ago",
							i, client.ClientID, client.ClientName, lastHeartbeat)
					}
					if len(clients) > maxLog {
						logger.Printf("  ... and %d more clients", len(clients)-maxLog)
					}
				}
			case <-metricsCtx.Done():
				return
			}
		}
	}()
	logger.Println("Started metrics logging goroutine (logs every minute)")
	
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Printf("gRPC server listening on :%d", *port)
		if err := s.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve: %w", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	select {
	case err := <-errChan:
		logger.Fatalf("Server error: %v", err)
	case sig := <-sigChan:
		logger.Printf("Received signal: %v, shutting down gracefully...", sig)
		
		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// Set health status to NOT_SERVING
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		
		// Graceful stop
		stopped := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(stopped)
		}()
		
		select {
		case <-stopped:
			logger.Println("Server stopped gracefully")
		case <-ctx.Done():
			logger.Println("Graceful shutdown timeout, forcing stop...")
			s.Stop()
		}
	}
}

