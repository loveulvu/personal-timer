package memories

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListMemoryEvidenceRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/memories/:id/evidence", NewHandler(nil, nil).ListMemoryEvidence)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/api/memories/0/evidence", nil))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.Code)
	}
}
