package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/amq"
	"anirank/api/internal/usecase/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type AMQHandler struct {
	lobbyManager *amq.LobbyManager
	jwtService   *auth.JWTService
	userRepo     domain.UserRepository
}

func NewAMQHandler(lobbyManager *amq.LobbyManager, jwtService *auth.JWTService, userRepo domain.UserRepository) *AMQHandler {
	return &AMQHandler{
		lobbyManager: lobbyManager,
		jwtService:   jwtService,
		userRepo:     userRepo,
	}
}

type wsConnWrapper struct {
	conn *websocket.Conn
}

func (w *wsConnWrapper) WriteJSON(v interface{}) error {
	return w.conn.WriteJSON(v)
}

func (w *wsConnWrapper) Close() error {
	return w.conn.Close()
}

// CreateRoom handles POST /api/amq/rooms
func (h *AMQHandler) CreateRoom(c *fiber.Ctx) error {
	var body struct {
		Config        domain.AMQConfig `json:"config"`
		GuestNickname string           `json:"guest_nickname"`
		GuestDeviceID string           `json:"guest_device_id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return domain.NewAppError(http.StatusBadRequest, "Invalid body", err)
	}

	// Try to get authenticated user from context (via optional auth middleware)
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

// ListRooms handles GET /api/amq/rooms
func (h *AMQHandler) ListRooms(c *fiber.Ctx) error {
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
func (h *AMQHandler) WSUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		roomID := c.Params("roomID")
		token := c.Query("token")
		nickname := c.Query("nickname", "Guest")
		deviceID := c.Query("device_id")
		spectator := c.Query("spectator")

		c.Locals("roomID", roomID)
		c.Locals("token", token)
		c.Locals("nickname", nickname)
		c.Locals("deviceID", deviceID)
		c.Locals("spectator", spectator == "true")

		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WSHandler runs the websocket reader loop after upgrade
func (h *AMQHandler) WSHandler(c *websocket.Conn) {
	roomIDVal := c.Locals("roomID")
	tokenVal := c.Locals("token")
	nicknameVal := c.Locals("nickname")
	deviceIDVal := c.Locals("deviceID")
	spectatorVal := c.Locals("spectator")

	deviceID := ""
	if deviceIDVal != nil {
		deviceID = deviceIDVal.(string)
	}

	if roomIDVal == nil || deviceIDVal == nil || deviceID == "" {
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
		if err == nil && claims != nil {
			u, errUser := h.userRepo.GetByUUID(context.Background(), claims.UserUUID)
			if errUser == nil {
				user = u
				nickname = u.Name
			}
		}
	}

	sessionID := uuid.New().String()
	wrapper := &wsConnWrapper{conn: c}

	// Join the room
	err := h.lobbyManager.JoinRoom(roomID, sessionID, wrapper, user, nickname, deviceID, asSpectator)
	if err != nil {
		_ = c.WriteJSON(fiber.Map{"type": "error", "payload": err.Error()})
		_ = c.Close()
		return
	}

	defer func() {
		h.lobbyManager.LeaveRoom(roomID, sessionID)
	}()

	// Reader loop
	for {
		var msg struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := c.ReadJSON(&msg); err != nil {
			break // Connection closed or error
		}

		room := h.lobbyManager.GetRoom(roomID)
		if room == nil {
			break
		}

		switch msg.Type {
		case "player_ready_toggle":
			room.SendEvent(amq.RoomEvent{Type: amq.EvReady, Data: sessionID})
		case "submit_guess":
			var guessPayload struct {
				AnimeSlug string `json:"anime_slug"`
			}
			if err := json.Unmarshal(msg.Payload, &guessPayload); err == nil {
				room.SendEvent(amq.RoomEvent{Type: amq.EvSubmitGuess, Data: &amq.GuessEvent{
					SessionID: sessionID,
					AnimeSlug: guessPayload.AnimeSlug,
				}})
			}
		case "update_lobby_config":
			var configPayload domain.AMQConfig
			if err := json.Unmarshal(msg.Payload, &configPayload); err == nil {
				room.SendEvent(amq.RoomEvent{Type: amq.EvConfigUpdate, Data: &amq.ConfigUpdateEvent{
					SessionID: sessionID,
					Config:    configPayload,
				}})
			}
		case "start_game":
			room.StartGame(sessionID)
		case "skip_summary":
			room.SendEvent(amq.RoomEvent{Type: amq.EvSkipSummary, Data: sessionID})
		case "reset_to_lobby":
			room.SendEvent(amq.RoomEvent{Type: amq.EvResetToLobby, Data: sessionID})
		case "transfer_host":
			var payload struct {
				TargetSessionID string `json:"target_session_id"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil && payload.TargetSessionID != "" {
				room.SendEvent(amq.RoomEvent{Type: amq.EvTransferHost, Data: &amq.TransferHostEvent{
					SessionID:       sessionID,
					TargetSessionID: payload.TargetSessionID,
				}})
			}
		case "send_chat_message":
			var chatPayload struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(msg.Payload, &chatPayload); err == nil && chatPayload.Text != "" {
				room.SendEvent(amq.RoomEvent{Type: amq.EvChat, Data: &amq.ChatEvent{
					SessionID: sessionID,
					Text:      chatPayload.Text,
				}})
			}
		}
	}
}
