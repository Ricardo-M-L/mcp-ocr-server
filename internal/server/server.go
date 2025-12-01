package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ricardo/mcp-ocr-server/internal/config"
	"github.com/ricardo/mcp-ocr-server/internal/tools"
	"github.com/ricardo/mcp-ocr-server/pkg/logger"
	"go.uber.org/zap"
)

// Server MCP OCR Server
type Server struct {
	mcpServer   *mcp.Server
	toolHandler *tools.Handler
	config      *config.Config
}

// New 创建 MCP Server
func New(cfg *config.Config) (*Server, error) {
	// 创建 Tool Handler
	toolHandler, err := tools.NewHandler(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool handler: %w", err)
	}

	// 创建 MCP Server
	mcpServer := mcp.NewServer(&mcp.ServerOptions{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	})

	server := &Server{
		mcpServer:   mcpServer,
		toolHandler: toolHandler,
		config:      cfg,
	}

	// 注册处理器
	server.registerHandlers()

	logger.Info("MCP Server created",
		zap.String("name", cfg.Server.Name),
		zap.String("version", cfg.Server.Version),
	)

	return server, nil
}

// registerHandlers 注册处理器
func (s *Server) registerHandlers() {
	// 注册工具列表处理器
	s.mcpServer.HandleListTools(func(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
		logger.Debug("ListTools called")

		toolSchemas := tools.GetToolSchemas()

		return &mcp.ListToolsResult{
			Tools: toolSchemas,
		}, nil
	})

	// 注册工具调用处理器
	s.mcpServer.HandleCallTool(func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Info("CallTool called",
			zap.String("tool", request.Params.Name),
		)

		// 解析参数
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			arguments = make(map[string]interface{})
		}

		// 调用工具处理器
		result, err := s.toolHandler.Handle(ctx, request.Params.Name, arguments)
		if err != nil {
			logger.Error("Tool execution failed",
				zap.String("tool", request.Params.Name),
				zap.Error(err),
			)
			return nil, err
		}

		return result, nil
	})

	logger.Debug("Handlers registered")
}

// Start 启动服务器
func (s *Server) Start() error {
	logger.Info("Starting MCP Server...")

	// 使用标准输入输出进行通信
	if err := s.mcpServer.ServeStdio(); err != nil {
		return fmt.Errorf("failed to serve stdio: %w", err)
	}

	return nil
}

// Close 关闭服务器
func (s *Server) Close() error {
	logger.Info("Closing MCP Server...")

	if s.toolHandler != nil {
		if err := s.toolHandler.Close(); err != nil {
			logger.Error("Failed to close tool handler", zap.Error(err))
		}
	}

	logger.Info("MCP Server closed")
	return nil
}

// GetServer 获取 MCP Server 实例
func (s *Server) GetServer() *mcp.Server {
	return s.mcpServer
}