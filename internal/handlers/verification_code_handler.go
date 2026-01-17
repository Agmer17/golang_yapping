package handlers

import (
	"github.com/Agmer17/golang_yapping/internal/service"
	"github.com/gin-gonic/gin"
)

type VerificationHandler struct {
	VerifcationService service.VerificationServiceInterface
}

func NewVerificationHandler(svc *service.VerificationService) *VerificationHandler {

	return &VerificationHandler{
		VerifcationService: svc,
	}
}

func (h *VerificationHandler) RegisterRoutes(rg *gin.RouterGroup) {

}
