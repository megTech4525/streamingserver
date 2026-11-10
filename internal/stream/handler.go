package stream

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

//POST /STREAMS (LOGGED-IN USER CREATES A STREAM AND GETS THE SECRET KEY ONCE)

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("user_id") // set by auth Middleware

	st, err := h.svc.CreateStream(c.Request.Context(), userID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create stream"})
		return
	}
	//key is hidden from JSON elswhere (json:"-") so it was added here on purpose

	c.JSON(http.StatusCreated, gin.H{"message": st.ID})
}

//GET /STREAMS? STATUS=LIVE (PUBLIC: THE HOME PAGE LIST)

func (h *Handler) ListLive(c *gin.Context) {
	streams, err := h.svc.ListLive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load streams"})
		return
	}
	c.JSON(http.StatusOK, streams)
}

//post /internal/streams/validate (called by the media server only)

func (h *Handler) Validate(c *gin.Context) {
	var req struct {
		Key string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key required"})
		return
	}

	st, err := h.svc.ValidateKey(c.Request.Context(), req.Key)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
		return
	}
	if err := h.svc.MarkLive(c.Request.Context(), st.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not mark live"})
		return
	}
	//CHECK THIS OUT PUT BEFORE PRODUCTION TO KNOW WHAT IT RETURNS
	c.JSON(http.StatusOK, gin.H{"stream_id": st.ID, "playback_id": st.PlaybackID})
}
