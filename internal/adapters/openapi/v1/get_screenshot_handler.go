package v1

import (
	"context"
	"errors"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/dariomba/screen-go/internal/application/usecase"
	"github.com/dariomba/screen-go/internal/domain"
)

type GetScreenshotHandler struct {
	getScreenshotUseCase *usecase.GetScreenshot
}

func NewGetScreenshotHandler(getScreenshotUseCase *usecase.GetScreenshot) *GetScreenshotHandler {
	return &GetScreenshotHandler{
		getScreenshotUseCase: getScreenshotUseCase,
	}
}

func (uc *GetScreenshotHandler) Execute(ctx context.Context, request openapi.GetScreenshotRequestObject) (openapi.GetScreenshotResponseObject, error) {
	result, err := uc.getScreenshotUseCase.Execute(ctx, request.ID)
	if err != nil {
		if errors.Is(err, domain.ErrScreenshotNotFound) {
			return openapi.GetScreenshot404JSONResponse{
				Error: "screenshot not found",
			}, nil
		}
		return nil, err
	}

	headers := openapi.GetScreenshot200ResponseHeaders{
		ContentType:   result.ContentType,
		ContentLength: int(result.Size),
	}

	if result.ContentType == "application/pdf" {
		return openapi.GetScreenshot200ApplicationPdfResponse{
			Body:          result.Data,
			ContentLength: result.Size,
			Headers:       headers,
		}, nil
	}

	return openapi.GetScreenshot200ImagePngResponse{
		Body:          result.Data,
		ContentLength: result.Size,
		Headers:       headers,
	}, nil
}
