package usecase

import (
	"context"
	"fmt"

	"github.com/lt-tunesi/TpFinalIng/internal/domain"
)

type AnalyticsUsecase struct {
	analyticsRepo domain.AnalyticsRepository
}

func NewAnalyticsUsecase(analyticsRepo domain.AnalyticsRepository) *AnalyticsUsecase {
	return &AnalyticsUsecase{analyticsRepo: analyticsRepo}
}

func (uc *AnalyticsUsecase) GetAnalyticsReport(ctx context.Context) ([]byte, error) {
	report, err := uc.analyticsRepo.GetTravelTrends(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener reporte analítico desde la réplica: %w", err)
	}
	return report, nil
}