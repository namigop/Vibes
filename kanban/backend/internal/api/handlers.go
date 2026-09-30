package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// maxBodyBytes caps request bodies. A card is tiny; anything larger is a bug
// or an attack, and either way it should not be buffered.
const maxBodyBytes = 1 << 20 // 1 MiB

// decodeJSON reads a JSON body into dst, rejecting malformed or oversized
// payloads with the standard validation error.
//
// Unknown fields are tolerated on purpose: an older client sending extra keys
// should not start failing against a newer server.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if r.Body == nil {
		return invalid("request body is required")
	}

	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return invalid("request body is too large")
		}
		return invalid("request body must be valid JSON")
	}

	// Reject trailing content so `{"a":1}{"b":2}` is not silently accepted.
	if err := dec.Decode(new(struct{})); !errors.Is(err, io.EOF) {
		return invalid("request body must contain a single JSON object")
	}

	return nil
}

// pathUUID reads a path wildcard and validates it.
func pathUUID(r *http.Request, name, what string) (uuid.UUID, error) {
	return parseUUID(r.PathValue(name), what)
}

// ---- health ---------------------------------------------------------------

func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	if err := s.Health(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, HealthResponse{Status: "degraded", DB: "unreachable"})
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok", DB: "ok"})
}

// ---- boards ---------------------------------------------------------------

func (s *Service) handleListBoards(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	boards, err := s.ListBoards(ctx)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, boards)
}

func (s *Service) handleCreateBoard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	var req CreateBoardRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	board, err := s.CreateBoard(ctx, req.Name)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, board)
}

func (s *Service) handleGetBoard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "boardId", "board id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	detail, err := s.GetBoardDetail(ctx, id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Service) handleRenameBoard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "boardId", "board id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req RenameBoardRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	board, err := s.RenameBoard(ctx, id, req.Name)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, board)
}

func (s *Service) handleDeleteBoard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "boardId", "board id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := s.DeleteBoard(ctx, id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- columns --------------------------------------------------------------

func (s *Service) handleCreateColumn(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	boardID, err := pathUUID(r, "boardId", "board id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req CreateColumnRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	col, err := s.CreateColumn(ctx, boardID, req.Name)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, col)
}

func (s *Service) handleRenameColumn(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "columnId", "column id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req RenameColumnRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	col, err := s.RenameColumn(ctx, id, req.Name)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, col)
}

func (s *Service) handleMoveColumn(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "columnId", "column id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req MoveColumnRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	columns, err := s.MoveColumn(ctx, id, req.Position)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, columns)
}

func (s *Service) handleDeleteColumn(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "columnId", "column id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := s.DeleteColumn(ctx, id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- cards ----------------------------------------------------------------

func (s *Service) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	columnID, err := pathUUID(r, "columnId", "column id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req CardRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	card, err := s.CreateCard(ctx, columnID, req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, card)
}

func (s *Service) handleUpdateCard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "cardId", "card id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req CardRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	card, err := s.UpdateCard(ctx, id, req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, card)
}

func (s *Service) handleMoveCard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "cardId", "card id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	var req MoveCardRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	columnID, err := parseUUID(req.ColumnID, "column_id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	card, err := s.MoveCard(ctx, id, columnID, req.Position)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, card)
}

func (s *Service) handleDeleteCard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	id, err := pathUUID(r, "cardId", "card id")
	if err != nil {
		writeError(w, r, err)
		return
	}

	if err := s.DeleteCard(ctx, id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// requestContext bounds handler work so a stuck query cannot pin a connection
// for the lifetime of the client.
func requestContext(r *http.Request) (ctxT, func()) {
	return withTimeout(r, 15*time.Second)
}
