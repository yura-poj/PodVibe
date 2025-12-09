package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"podvibe/internal/auth"
	"podvibe/internal/models"
	"podvibe/internal/services"
)

type stubRecommendationService struct {
	items       []services.RecommendationItem
	total       int64
	err         error
	lastUserID  uint
	lastPage    int
	lastPerPage int
}

func (s *stubRecommendationService) Recommend(userID uint, page, pageSize int) ([]services.RecommendationItem, int64, error) {
	s.lastUserID = userID
	s.lastPage = page
	s.lastPerPage = pageSize
	return s.items, s.total, s.err
}

func TestRecommendationHandler_Response(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &stubRecommendationService{
		items: []services.RecommendationItem{
			{
				Episode: models.Episode{ID: 1, Title: "A"},
				Podcast: models.Podcast{ID: 10, Title: "Pod"},
				Author:  models.User{ID: 2, Username: "u"},
				Reason:  "from_follow",
			},
		},
		total: 1,
	}
	h := NewRecommendationHandler(stub)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/recommendations?page=2&page_size=5", nil)
	c.Request = req
	c.Set(auth.ContextUserID, uint(42))

	h.Recommend(c)

	if stub.lastUserID != 42 || stub.lastPage != 2 || stub.lastPerPage != 5 {
		t.Fatalf("service called with wrong params: user=%d page=%d size=%d", stub.lastUserID, stub.lastPage, stub.lastPerPage)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp struct {
		Items []struct {
			Episode models.Episode `json:"episode"`
			Author  struct {
				ID       uint   `json:"id"`
				Username string `json:"username"`
			} `json:"author"`
			Podcast struct {
				ID    uint   `json:"id"`
				Title string `json:"title"`
			} `json:"podcast"`
			Reason string `json:"reason"`
		} `json:"items"`
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
		Total    int64 `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Page != 2 || resp.PageSize != 5 || resp.Total != 1 {
		t.Fatalf("unexpected pagination in response: %+v", resp)
	}
	if len(resp.Items) != 1 || resp.Items[0].Episode.ID != 1 || resp.Items[0].Reason != "from_follow" {
		t.Fatalf("unexpected items: %+v", resp.Items)
	}
}
