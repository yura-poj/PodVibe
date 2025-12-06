package handlers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hajimehoshi/go-mp3"

	"podvibe/internal/auth"
	"podvibe/internal/models"
	"podvibe/internal/services"
	"podvibe/internal/storage"
)

type EpisodeHandler struct {
	episodes  *services.EpisodeService
	podcasts  *services.PodcastService
	storage   *storage.LocalStorage
	jwtSecret string
}

func NewEpisodeHandler(episodes *services.EpisodeService, podcasts *services.PodcastService, storage *storage.LocalStorage, jwtSecret string) *EpisodeHandler {
	return &EpisodeHandler{episodes: episodes, podcasts: podcasts, storage: storage, jwtSecret: jwtSecret}
}

func (h *EpisodeHandler) Create(c *gin.Context) {
	podcastID, _ := strconv.Atoi(c.Param("id"))
	title := c.PostForm("title")
	if title == "" {
		RespondError(c, http.StatusBadRequest, "validation_error", "title is required")
		return
	}
	uid := auth.UserID(c)
	podcast, err := h.podcasts.Get(uint(podcastID))
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "podcast not found")
		return
	}
	file, err := c.FormFile("audio")
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", "audio file is required")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExt := map[string]bool{".mp3": true, ".wav": true}
	if !allowedExt[ext] {
		RespondError(c, http.StatusBadRequest, "validation_error", "unsupported audio format (allowed: mp3, wav)")
		return
	}
	subPath := fmt.Sprintf("audio/%d/%d_%d%s", uid, podcastID, time.Now().UnixNano(), ext)
	f, _ := file.Open()
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", "failed to read audio file")
		return
	}
	dur, err := audioDuration(ext, data)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if dur > 3*time.Minute {
		RespondError(c, http.StatusBadRequest, "validation_error", "audio must be shorter than 3 minutes")
		return
	}
	audioPath, err := h.storage.Save(subPath, bytes.NewReader(data))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	tags := parseTags(c.PostForm("tags"))
	ep, err := h.episodes.Create(c, podcast.OwnerID, uid, podcast.ID, title, c.PostForm("description"), audioPath, tags)
	if err != nil {
		RespondError(c, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	c.JSON(http.StatusCreated, h.serializeEpisode(ep))
}

func parseTags(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == '#' || r == ' ' })
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (h *EpisodeHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ep, err := h.episodes.Get(uint(id))
	if err != nil {
		RespondError(c, http.StatusNotFound, "not_found", "episode not found")
		return
	}
	c.JSON(http.StatusOK, h.serializeEpisode(ep))
}

func audioDuration(ext string, data []byte) (time.Duration, error) {
	switch ext {
	case ".mp3":
		dec, err := mp3.NewDecoder(bytes.NewReader(data))
		if err != nil {
			return 0, fmt.Errorf("cannot decode mp3")
		}
		sr := dec.SampleRate()
		if sr == 0 {
			return 0, fmt.Errorf("invalid mp3 sample rate")
		}
		length := dec.Length()
		if length == 0 {
			buf := make([]byte, 2048)
			var total int64
			for {
				n, err := dec.Read(buf)
				total += int64(n)
				if err == io.EOF {
					break
				}
				if err != nil {
					return 0, fmt.Errorf("cannot read mp3")
				}
			}
			length = total
		}
		seconds := float64(length) / float64(sr*4) // 2 channels * 2 bytes per sample
		return time.Duration(seconds * float64(time.Second)), nil
	case ".wav":
		dur, err := parseWAVDuration(data)
		if err != nil {
			return 0, fmt.Errorf("cannot read wav: %w", err)
		}
		return dur, nil
	default:
		return 0, fmt.Errorf("unsupported audio format")
	}
}

func parseWAVDuration(data []byte) (time.Duration, error) {
	if len(data) < 44 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return 0, fmt.Errorf("invalid wav header")
	}
	var byteRate uint32
	pos := 12
	for pos+8 <= len(data) {
		chunkID := string(data[pos : pos+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		pos += 8
		if pos+chunkSize > len(data) {
			return 0, fmt.Errorf("corrupt wav chunk")
		}
		switch chunkID {
		case "fmt ":
			if chunkSize < 16 {
				return 0, fmt.Errorf("invalid fmt chunk")
			}
			byteRate = binary.LittleEndian.Uint32(data[pos+8 : pos+12])
		case "data":
			if byteRate == 0 {
				return 0, fmt.Errorf("missing fmt chunk")
			}
			seconds := float64(chunkSize) / float64(byteRate)
			return time.Duration(seconds * float64(time.Second)), nil
		}
		pos += chunkSize
		if chunkSize%2 == 1 {
			pos++
		}
	}
	return 0, fmt.Errorf("data chunk not found")
}

func (h *EpisodeHandler) ListForPodcast(c *gin.Context) {
	podcastID, _ := strconv.Atoi(c.Param("id"))
	page, pageSize := parsePagination(c)
	list, total, err := h.episodes.ListByPodcast(uint(podcastID), page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for i := range list {
		items = append(items, h.serializeEpisode(&list[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h *EpisodeHandler) Play(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var userID uint
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if claims, err := auth.ParseToken(parts[1], h.jwtSecret); err == nil {
				userID = claims.UserID
			}
		}
	}
	if err := h.episodes.AddPlay(userID, uint(id)); err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusAccepted)
}

func (h *EpisodeHandler) Popular(c *gin.Context) {
	page, pageSize := parsePagination(c)
	list, total, err := h.episodes.Popular(page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	items := make([]gin.H, 0, len(list))
	for i := range list {
		items = append(items, h.serializeEpisode(&list[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page, "page_size": pageSize, "total": total})
}

func (h *EpisodeHandler) serializeEpisode(ep *models.Episode) gin.H {
	tags := []string{}
	for _, t := range ep.Tags {
		tags = append(tags, t.Tag)
	}
	return gin.H{
		"id":                ep.ID,
		"podcast_id":        ep.PodcastID,
		"title":             ep.Title,
		"description":       ep.Description,
		"audio_url":         h.storage.PublicPath(ep.AudioPath),
		"duration_seconds":  ep.DurationSeconds,
		"transcript":        ep.Transcript,
		"transcript_status": ep.TranscriptStatus,
		"like_count":        ep.LikeCount,
		"comment_count":     ep.CommentCount,
		"play_count":        ep.PlayCount,
		"tags":              tags,
		"published_at":      ep.PublishedAt,
	}
}
