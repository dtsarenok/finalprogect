package calculator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"distributed-calculator/internal/auth"

	"github.com/Knetic/govaluate"
)

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	Result string `json:"result"`
}

func CalculateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CalculateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		expression, err := govaluate.NewEvaluableExpression(req.Expression)
		if err != nil {
			http.Error(w, "invalid expression", http.StatusBadRequest)
			return
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			http.Error(w, "evaluation error", http.StatusBadRequest)
			return
		}
		resultStr := fmt.Sprintf("%v", result)

		_, err = db.Exec(
			"INSERT INTO calculations (user_id, expression, result, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)",
			userID, req.Expression, resultStr,
		)
		if err != nil {
			http.Error(w, "failed to save calculation", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(CalculateResponse{Result: resultStr})
	}
}
