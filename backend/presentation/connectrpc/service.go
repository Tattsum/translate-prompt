package connectrpc

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	domainintake "github.com/Tattsum/translate-prompt/backend/domain/intake"
	translatepromptv1 "github.com/Tattsum/translate-prompt/backend/gen/translate_prompt/v1"
	"github.com/Tattsum/translate-prompt/backend/gen/translate_prompt/v1/translate_promptv1connect"
	"github.com/Tattsum/translate-prompt/backend/presentation/mapper"
)

// Service implements TranslatePromptService via Connect-RPC.
type Service struct {
	optimize           *optimize.UseCase
	intake             *appintake.UseCase
	investigateEnabled bool
}

// NewService wires use cases into the Connect handler.
func NewService(opt *optimize.UseCase, intake *appintake.UseCase, investigateEnabled bool) *Service {
	return &Service{
		optimize:           opt,
		intake:             intake,
		investigateEnabled: investigateEnabled,
	}
}

// Mount registers the Connect service on mux at the standard path prefix.
func Mount(mux *http.ServeMux, s *Service) {
	path, handler := translate_promptv1connect.NewTranslatePromptServiceHandler(s)
	mux.Handle(path, handler)
}

func (s *Service) Health(
	_ context.Context,
	_ *connect.Request[translatepromptv1.HealthRequest],
) (*connect.Response[translatepromptv1.HealthResponse], error) {
	return connect.NewResponse(&translatepromptv1.HealthResponse{Status: "ok"}), nil
}

func (s *Service) Analyze(
	ctx context.Context,
	req *connect.Request[translatepromptv1.AnalyzeRequest],
) (*connect.Response[translatepromptv1.AnalyzeResponse], error) {
	cfg := mapper.ConfigFromProto(req.Msg.Config)
	result, err := s.intake.Analyze(ctx, req.Msg.Prompt, cfg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("analyze: %w", err))
	}
	return connect.NewResponse(mapper.AnalyzeToProto(result)), nil
}

func (s *Service) Investigate(
	ctx context.Context,
	req *connect.Request[translatepromptv1.InvestigateRequest],
) (*connect.Response[translatepromptv1.InvestigateResponse], error) {
	if !s.investigateEnabled {
		return nil, connect.NewError(connect.CodePermissionDenied, domainintake.ErrInvestigateDisabled)
	}
	profile := mapper.ProfileFromString(req.Msg.TargetProfile)
	result, err := s.intake.Investigate(ctx, req.Msg.WorkspacePath, profile)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("investigate: %w", err))
	}
	return connect.NewResponse(mapper.InvestigateToProto(result)), nil
}

func (s *Service) Optimize(
	ctx context.Context,
	req *connect.Request[translatepromptv1.OptimizeRequest],
) (*connect.Response[translatepromptv1.OptimizeResponse], error) {
	cfg := mapper.ConfigFromProto(req.Msg.Config)
	promptText, cfg, err := mapper.PrepareOptimizePrompt(ctx, s.intake, req.Msg.Prompt, cfg, req.Msg.Answers)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("prepare: %w", err))
	}
	result, err := s.optimize.Optimize(ctx, promptText, cfg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("optimize: %w", err))
	}
	return connect.NewResponse(mapper.OptimizeToProto(result)), nil
}

func (s *Service) Estimate(
	_ context.Context,
	req *connect.Request[translatepromptv1.EstimateRequest],
) (*connect.Response[translatepromptv1.EstimateResponse], error) {
	tokens, err := s.optimize.Estimate(req.Msg.Text, req.Msg.Tokenizer)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("estimate: %w", err))
	}
	return connect.NewResponse(&translatepromptv1.EstimateResponse{Tokens: int32(tokens)}), nil
}

var _ translate_promptv1connect.TranslatePromptServiceHandler = (*Service)(nil)
