package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"main.go/internal/auth"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.RegisterUser(c.Request.Context(), req.Username, req.Email, req.Password)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "could not register(email or username)"})
		return

	}
	c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	token, err := auth.GenerateToken(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
