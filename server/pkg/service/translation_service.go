package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/dasmlab/nanabush/server/pkg/proto/v1"
)

// ClientInfo tracks registered client information.
type ClientInfo struct {
	ClientID    string
	ClientName  string
	ClientVersion string
	Namespace   string
	Metadata    map[string]string
	RegisteredAt time.Time
	LastHeartbeat time.Time
}

// TranslationService implements the TranslationService gRPC service.
type TranslationService struct {
	nanabushv1.UnimplementedTranslationServiceServer
	
	// Backend is the vLLM backend integration (to be implemented)
	Backend TranslatorBackend
	
	// Logger for service operations
	Logger *log.Logger
	
	// Client tracking
	clients      map[string]*ClientInfo
	clientsMutex sync.RWMutex
	clientIDCounter int64
	heartbeatInterval int32 // seconds
}

// TranslatorBackend defines the interface for vLLM backend integration.
type TranslatorBackend interface {
	// TranslateTitle translates a title only (lightweight operation)
	TranslateTitle(ctx context.Context, title, sourceLang, targetLang string) (string, error)
	
	// TranslateDocument translates a full document
	TranslateDocument(ctx context.Context, doc *nanabushv1.DocumentContent, sourceLang, targetLang string) (*nanabushv1.DocumentContent, error)
	
	// CheckHealth checks if the backend is ready
	CheckHealth(ctx context.Context) error
}

// NewTranslationService creates a new TranslationService instance.
func NewTranslationService(backend TranslatorBackend, logger *log.Logger) *TranslationService {
	if logger == nil {
		logger = log.Default()
	}
	return &TranslationService{
		Backend:          backend,
		Logger:           logger,
		clients:          make(map[string]*ClientInfo),
		heartbeatInterval: 60, // Default: 60 seconds
	}
}

// RegisterClient registers a new client with the server.
func (s *TranslationService) RegisterClient(ctx context.Context, req *nanabushv1.RegisterClientRequest) (*nanabushv1.RegisterClientResponse, error) {
	s.Logger.Printf("RegisterClient request: name=%q, version=%q, namespace=%q", req.ClientName, req.ClientVersion, req.Namespace)
	
	// Validate request
	if req.ClientName == "" {
		return nil, status.Error(codes.InvalidArgument, "client_name is required")
	}
	
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	
	// Generate unique client ID
	s.clientIDCounter++
	clientID := fmt.Sprintf("client-%d-%d", time.Now().Unix(), s.clientIDCounter)
	
	now := time.Now()
	// Create client info
	clientInfo := &ClientInfo{
		ClientID:      clientID,
		ClientName:    req.ClientName,
		ClientVersion: req.ClientVersion,
		Namespace:     req.Namespace,
		Metadata:      req.Metadata,
		RegisteredAt:  now,
		LastHeartbeat: now,
	}
	
	// Store client
	s.clients[clientID] = clientInfo
	
	s.Logger.Printf("Client registered: id=%q, name=%q, total_clients=%d", clientID, req.ClientName, len(s.clients))
	
	// Calculate expiration (24 hours from now)
	expiresAt := now.Add(24 * time.Hour)
	
	return &nanabushv1.RegisterClientResponse{
		ClientId:               clientID,
		Success:                true,
		Message:                fmt.Sprintf("Client %q registered successfully", req.ClientName),
		HeartbeatIntervalSeconds: s.heartbeatInterval,
		ExpiresAt:              timestamppb.New(expiresAt),
	}, nil
}

// Heartbeat sends a keepalive and re-authentication signal from the client.
func (s *TranslationService) Heartbeat(ctx context.Context, req *nanabushv1.HeartbeatRequest) (*nanabushv1.HeartbeatResponse, error) {
	s.Logger.Printf("Heartbeat request: client_id=%q, client_name=%q", req.ClientId, req.ClientName)
	
	// Validate request
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "client_id is required")
	}
	if req.ClientName == "" {
		return nil, status.Error(codes.InvalidArgument, "client_name is required")
	}
	
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	
	// Look up client
	clientInfo, exists := s.clients[req.ClientId]
	if !exists {
		s.Logger.Printf("Heartbeat from unknown client: id=%q, name=%q", req.ClientId, req.ClientName)
		return &nanabushv1.HeartbeatResponse{
			Success:             false,
			Message:             "Client not registered or expired",
			ReceivedAt:          timestamppb.Now(),
			HeartbeatIntervalSeconds: s.heartbeatInterval,
			ReRegisterRequired: true,
		}, nil
	}
	
	// Validate client name matches
	if clientInfo.ClientName != req.ClientName {
		s.Logger.Printf("Heartbeat client name mismatch: expected=%q, got=%q", clientInfo.ClientName, req.ClientName)
		return &nanabushv1.HeartbeatResponse{
			Success:             false,
			Message:             "Client name mismatch",
			ReceivedAt:          timestamppb.Now(),
			HeartbeatIntervalSeconds: s.heartbeatInterval,
			ReRegisterRequired: true,
		}, nil
	}
	
	// Update last heartbeat time
	clientInfo.LastHeartbeat = time.Now()
	
	// Check if registration expired (24 hours)
	if time.Since(clientInfo.RegisteredAt) > 24*time.Hour {
		s.Logger.Printf("Client registration expired: id=%q, name=%q", req.ClientId, req.ClientName)
		delete(s.clients, req.ClientId)
		return &nanabushv1.HeartbeatResponse{
			Success:             false,
			Message:             "Registration expired",
			ReceivedAt:          timestamppb.Now(),
			HeartbeatIntervalSeconds: s.heartbeatInterval,
			ReRegisterRequired: true,
		}, nil
	}
	
	s.Logger.Printf("Heartbeat acknowledged: client_id=%q, name=%q, last_seen=%v", 
		req.ClientId, req.ClientName, clientInfo.LastHeartbeat)
	
	return &nanabushv1.HeartbeatResponse{
		Success:             true,
		Message:             "Heartbeat acknowledged",
		ReceivedAt:          timestamppb.Now(),
		HeartbeatIntervalSeconds: s.heartbeatInterval,
		ReRegisterRequired: false,
	}, nil
}

// GetRegisteredClients returns all currently registered clients (for monitoring/debugging).
func (s *TranslationService) GetRegisteredClients() []*ClientInfo {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	clients := make([]*ClientInfo, 0, len(s.clients))
	for _, client := range s.clients {
		// Create a copy to avoid race conditions
		clientCopy := *client
		clients = append(clients, &clientCopy)
	}
	return clients
}

// GetClientCount returns the number of currently registered clients.
func (s *TranslationService) GetClientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return len(s.clients)
}

// GetClientMetrics returns metrics about registered clients.
type ClientMetrics struct {
	TotalClients      int
	ClientsByNamespace map[string]int
	ClientsByVersion   map[string]int
	OldestHeartbeat    time.Time
	NewestHeartbeat    time.Time
}

// GetClientMetrics returns aggregated metrics about registered clients.
func (s *TranslationService) GetClientMetrics() *ClientMetrics {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	metrics := &ClientMetrics{
		TotalClients:       len(s.clients),
		ClientsByNamespace: make(map[string]int),
		ClientsByVersion:   make(map[string]int),
	}
	
	if len(s.clients) == 0 {
		return metrics
	}
	
	now := time.Now()
	oldest := now
	newest := time.Time{}
	
	for _, client := range s.clients {
		// Count by namespace
		ns := client.Namespace
		if ns == "" {
			ns = "unknown"
		}
		metrics.ClientsByNamespace[ns]++
		
		// Count by version
		version := client.ClientVersion
		if version == "" {
			version = "unknown"
		}
		metrics.ClientsByVersion[version]++
		
		// Track heartbeat times
		if client.LastHeartbeat.Before(oldest) {
			oldest = client.LastHeartbeat
		}
		if client.LastHeartbeat.After(newest) {
			newest = client.LastHeartbeat
		}
	}
	
	metrics.OldestHeartbeat = oldest
	metrics.NewestHeartbeat = newest
	
	return metrics
}

// CleanupExpiredClients removes clients that haven't sent a heartbeat in a while.
// This should be called periodically (e.g., every 5 minutes).
func (s *TranslationService) CleanupExpiredClients(maxIdleTime time.Duration) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()
	
	now := time.Now()
	removed := 0
	
	for clientID, client := range s.clients {
		if now.Sub(client.LastHeartbeat) > maxIdleTime {
			s.Logger.Printf("Removing expired client: id=%q, name=%q, last_heartbeat=%v", 
				clientID, client.ClientName, client.LastHeartbeat)
			delete(s.clients, clientID)
			removed++
		}
	}
	
	if removed > 0 {
		s.Logger.Printf("Cleaned up %d expired clients, %d remaining", removed, len(s.clients))
	}
}

// CheckTitle performs a lightweight pre-flight check with title only.
// This validates that Nanabush is ready and can handle the request.
func (s *TranslationService) CheckTitle(ctx context.Context, req *nanabushv1.TitleCheckRequest) (*nanabushv1.TitleCheckResponse, error) {
	s.Logger.Printf("CheckTitle request: title=%q, source=%q, target=%q", req.Title, req.SourceLanguage, req.LanguageTag)
	
	// Validate request
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.LanguageTag == "" {
		return nil, status.Error(codes.InvalidArgument, "language_tag is required")
	}
	if req.SourceLanguage == "" {
		return nil, status.Error(codes.InvalidArgument, "source_language is required")
	}
	
	// Check backend health
	if s.Backend != nil {
		if err := s.Backend.CheckHealth(ctx); err != nil {
			s.Logger.Printf("Backend health check failed: %v", err)
			return &nanabushv1.TitleCheckResponse{
				Ready:                false,
				Message:              fmt.Sprintf("Backend not ready: %v", err),
				EstimatedTimeSeconds: 0,
			}, nil
		}
	}
	
	// Estimate time based on title length (simple heuristic)
	// TODO: Improve with actual backend metrics
	estimatedSeconds := int32(5 + len(req.Title)/10)
	if estimatedSeconds < 5 {
		estimatedSeconds = 5
	}
	if estimatedSeconds > 60 {
		estimatedSeconds = 60
	}
	
	s.Logger.Printf("CheckTitle response: ready=true, estimated=%ds", estimatedSeconds)
	
	return &nanabushv1.TitleCheckResponse{
		Ready:                true,
		Message:              "Ready to handle translation request",
		EstimatedTimeSeconds: estimatedSeconds,
	}, nil
}

// Translate performs full document translation.
// This is the main translation endpoint that processes complete documents.
func (s *TranslationService) Translate(ctx context.Context, req *nanabushv1.TranslateRequest) (*nanabushv1.TranslateResponse, error) {
	s.Logger.Printf("Translate request: job_id=%q, primitive=%v, namespace=%q", req.JobId, req.Primitive, req.Namespace)
	
	startTime := time.Now()
	
	// Validate request
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}
	if req.TargetLanguage == "" {
		return nil, status.Error(codes.InvalidArgument, "target_language is required")
	}
	if req.SourceLanguage == "" {
		return nil, status.Error(codes.InvalidArgument, "source_language is required")
	}
	
	sourceLang := req.SourceLanguage
	targetLang := req.TargetLanguage
	
	var translatedTitle string
	var translatedDoc *nanabushv1.DocumentContent
	var err error
	
	// Handle different primitive types
	switch req.Primitive {
	case nanabushv1.PrimitiveType_PRIMITIVE_TITLE:
		// Title-only translation
		if req.GetTitle() == "" {
			return nil, status.Error(codes.InvalidArgument, "title is required for PRIMITIVE_TITLE")
		}
		
		if s.Backend != nil {
			translatedTitle, err = s.Backend.TranslateTitle(ctx, req.GetTitle(), sourceLang, targetLang)
			if err != nil {
				s.Logger.Printf("Translate title failed: %v", err)
				return &nanabushv1.TranslateResponse{
					JobId:        req.JobId,
					Success:      false,
					ErrorMessage: fmt.Sprintf("Translation failed: %v", err),
					CompletedAt:  timestamppb.Now(),
				}, nil
			}
		} else {
			// Placeholder: return original title when backend not implemented
			translatedTitle = req.GetTitle() + " [translated]"
		}
		
	case nanabushv1.PrimitiveType_PRIMITIVE_DOC_TRANSLATE:
		// Full document translation
		if req.GetDoc() == nil {
			return nil, status.Error(codes.InvalidArgument, "doc is required for PRIMITIVE_DOC_TRANSLATE")
		}
		
		if s.Backend != nil {
			translatedDoc, err = s.Backend.TranslateDocument(ctx, req.GetDoc(), sourceLang, targetLang)
			if err != nil {
				s.Logger.Printf("Translate document failed: %v", err)
				return &nanabushv1.TranslateResponse{
					JobId:        req.JobId,
					Success:      false,
					ErrorMessage: fmt.Sprintf("Translation failed: %v", err),
					CompletedAt:  timestamppb.Now(),
				}, nil
			}
		} else {
			// Placeholder: return original document when backend not implemented
			translatedDoc = &nanabushv1.DocumentContent{
				Title:    req.GetDoc().Title + " [translated]",
				Markdown: req.GetDoc().Markdown + "\n\n*[Translated from " + sourceLang + " to " + targetLang + "]*",
				Slug:     req.GetDoc().Slug,
				Metadata: req.GetDoc().Metadata,
			}
		}
		
	default:
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unsupported primitive type: %v", req.Primitive))
	}
	
	// Build response
	inferenceTime := time.Since(startTime).Seconds()
	
	resp := &nanabushv1.TranslateResponse{
		JobId:               req.JobId,
		Success:             true,
		CompletedAt:         timestamppb.Now(),
		TokensUsed:          0, // TODO: Get from backend
		InferenceTimeSeconds: inferenceTime,
	}
	
	if translatedTitle != "" {
		resp.TranslatedTitle = translatedTitle
	}
	if translatedDoc != nil {
		resp.TranslatedMarkdown = translatedDoc.Markdown
		if translatedDoc.Title != "" {
			resp.TranslatedTitle = translatedDoc.Title
		}
	}
	
	s.Logger.Printf("Translate response: job_id=%q, success=true, time=%.2fs", req.JobId, inferenceTime)
	
	return resp, nil
}

// TranslateStream supports streaming for large documents.
// Client sends chunks, server responds with translated chunks.
func (s *TranslationService) TranslateStream(stream nanabushv1.TranslationService_TranslateStreamServer) error {
	s.Logger.Println("TranslateStream request started")
	
	var jobID string
	chunkIndex := int32(0)
	
	for {
		// Receive chunk from client
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				// Client closed stream
				break
			}
			s.Logger.Printf("TranslateStream receive error: %v", err)
			return status.Error(codes.Internal, fmt.Sprintf("failed to receive chunk: %v", err))
		}
		
		if jobID == "" {
			jobID = chunk.JobId
			s.Logger.Printf("TranslateStream started for job_id=%q", jobID)
		}
		
		// Check if this is the final chunk
		if chunk.IsFinal {
			s.Logger.Printf("TranslateStream final chunk received for job_id=%q", jobID)
			
			// Send final acknowledgment
			if err := stream.Send(&nanabushv1.TranslateChunk{
				JobId:      jobID,
				ChunkIndex: chunkIndex,
				IsFinal:    true,
				Content:    "[Stream completed]",
			}); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("failed to send final chunk: %v", err))
			}
			break
		}
		
		// TODO: Implement actual streaming translation
		// For now, echo back with translation placeholder
		translatedContent := chunk.Content + " [translated chunk " + fmt.Sprintf("%d", chunkIndex) + "]"
		
		// Send translated chunk back to client
		if err := stream.Send(&nanabushv1.TranslateChunk{
			JobId:      jobID,
			ChunkIndex: chunkIndex,
			IsFinal:    false,
			Content:    translatedContent,
		}); err != nil {
			s.Logger.Printf("TranslateStream send error: %v", err)
			return status.Error(codes.Internal, fmt.Sprintf("failed to send chunk: %v", err))
		}
		
		chunkIndex++
	}
	
	s.Logger.Printf("TranslateStream completed for job_id=%q", jobID)
	return nil
}

