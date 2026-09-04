package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/usecase/rate"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type RateHandler struct {
	lobbyManager *rate.LobbyManager
	jwtService   *auth.JWTService
	userRepo     domain.UserRepository
}

func NewRateHandler(lobbyManager *rate.LobbyManager, jwtService *auth.JWTService, userRepo domain.UserRepository) *RateHandler {
	return &RateHandler{
		lobbyManager: lobbyManager,
		jwtService:   jwtService,
		userRepo:     userRepo,
	}
}

// CreateRoom handles POST /api/rate/rooms
func (h *RateHandler) CreateRoom(c *fiber.Ctx) error {
	var body struct {
		Config        domain.RateConfig `json:"config"`
		GuestNickname string            `json:"guest_nickname"`
		GuestDeviceID string            `json:"guest_device_id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return domain.NewAppError(http.StatusBadRequest, "Invalid body", err)
	}

	var authUser *domain.User
	if val := c.Locals("user"); val != nil {
		if u, ok := val.(*domain.User); ok {
			authUser = u
		}
	}

	roomID, err := h.lobbyManager.CreateRoom(c.Context(), body.Config, authUser, body.GuestNickname, body.GuestDeviceID)
	if err != nil {
		return domain.NewAppError(http.StatusInternalServerError, "Could not create room", err)
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"room_id": roomID,
		},
	})
}

// ListRooms handles GET /api/rate/rooms
func (h *RateHandler) ListRooms(c *fiber.Ctx) error {
	rooms, err := h.lobbyManager.ListPublicRooms(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    rooms,
	})
}

// WSUpgrade handles upgrading HTTP connection to WebSocket
func (h *RateHandler) WSUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("roomID", c.Params("roomID"))
		c.Locals("token", c.Query("token"))
		c.Locals("nickname", c.Query("nickname", "Guest"))
		c.Locals("deviceID", c.Query("device_id"))
		c.Locals("spectator", c.Query("spectator") == "true")
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WSHandler runs the websocket reader loop after upgrade
func (h *RateHandler) WSHandler(c *websocket.Conn) {
	roomIDVal := c.Locals("roomID")
	tokenVal := c.Locals("token")
	nicknameVal := c.Locals("nickname")
	deviceIDVal := c.Locals("deviceID")
	spectatorVal := c.Locals("spectator")

	deviceID := ""
	if deviceIDVal != nil {
		deviceID = deviceIDVal.(string)
	}

	if roomIDVal == nil || deviceID == "" {
		_ = c.WriteJSON(fiber.Map{"type": "error", "payload": "Missing room_id or device_id"})
		_ = c.Close()
		return
	}

	roomID := roomIDVal.(string)
	token := ""
	if tokenVal != nil {
		token = tokenVal.(string)
	}
	nickname := "Guest"
	if nicknameVal != nil {
		nickname = nicknameVal.(string)
	}
	asSpectator := false
	if spectatorVal != nil {
		asSpectator = spectatorVal.(bool)
	}

	var user *domain.User
	if token != "" {
		claims, err := h.jwtService.ValidateToken(token)
		if err == nil && claims != nil && claims.UserUUID != "" {
			u, errUser := h.userRepo.GetByUUID(context.Background(), claims.UserUUID)
			if errUser == nil {
				user = u
				nickname = u.Name
			}
		}
	}

	sessionID := uuid.New().String()
	wrapper := &wsConnWrapper{conn: c}

	if err := h.lobbyManager.JoinRoom(roomID, sessionID, wrapper, user, nickname, deviceID, asSpectator); err != nil {
		_ = c.WriteJSON(fiber.Map{"type": "error", "payload": err.Error()})
		_ = c.Close()
		return
	}
	defer h.lobbyManager.LeaveRoom(roomID, sessionID)

	for {
		var msg struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := c.ReadJSON(&msg); err != nil {
			break
		}

		room := h.lobbyManager.GetRoom(roomID)
		if room == nil {
			break
		}

		switch msg.Type {
		case "update_lobby_config":
			var configPayload domain.RateConfig
			if err := json.Unmarshal(msg.Payload, &configPayload); err == nil {
				room.SendEvent(rate.RoomEvent{Type: rate.EvConfigUpdate, Data: &rate.ConfigUpdateEvent{
					SessionID: sessionID,
					Config:    configPayload,
				}})
			}
		case "start_session":
			room.SendEvent(rate.RoomEvent{Type: rate.EvStartSession, Data: sessionID})
		case "queue_add":
			var payload struct {
				SongUUID string `json:"song_uuid"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				room.SendEvent(rate.RoomEvent{Type: rate.EvQueueAdd, Data: &rate.QueueAddEvent{
					SessionID: sessionID,
					SongUUID:  payload.SongUUID,
				}})
			}
		case "queue_remove":
			var payload struct {
				ItemID string `json:"item_id"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				room.SendEvent(rate.RoomEvent{Type: rate.EvQueueRemove, Data: &rate.QueueRemoveEvent{
					SessionID: sessionID,
					ItemID:    payload.ItemID,
				}})
			}
		case "set_song":
			var payload struct {
				SongUUID string `json:"song_uuid"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				room.SendEvent(rate.RoomEvent{Type: rate.EvSetSong, Data: &rate.SetSongEvent{
					SessionID: sessionID,
					SongUUID:  payload.SongUUID,
				}})
			}
		case "next":
			room.SendEvent(rate.RoomEvent{Type: rate.EvNext, Data: sessionID})
		case "submit_rating":
			var payload struct {
				Score float64 `json:"score"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				room.SendEvent(rate.RoomEvent{Type: rate.EvSubmitRating, Data: &rate.RatingEvent{
					SessionID: sessionID,
					Score:     payload.Score,
				}})
			}
		case "end_session":
			room.SendEvent(rate.RoomEvent{Type: rate.EvEndSession, Data: sessionID})
		case "reset_to_lobby":
			room.SendEvent(rate.RoomEvent{Type: rate.EvResetToLobby, Data: sessionID})
		case "transfer_host":
			var payload struct {
				TargetSessionID string `json:"target_session_id"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				room.SendEvent(rate.RoomEvent{Type: rate.EvTransferHost, Data: &rate.TransferHostEvent{
					SessionID:       sessionID,
					TargetSessionID: payload.TargetSessionID,
				}})
			}
		case "send_chat_message":
			var chatPayload struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(msg.Payload, &chatPayload); err == nil && chatPayload.Text != "" {
				room.SendEvent(rate.RoomEvent{Type: rate.EvChat, Data: &rate.ChatEvent{
					SessionID: sessionID,
					Text:      chatPayload.Text,
				}})
			}
		case "close_room":
			room.SendEvent(rate.RoomEvent{Type: rate.EvCloseRoom, Data: sessionID})
		}
	}
}
