package broadcasturl

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type broadcastURLResponse struct {
	ID        *int       `json:"id" db:"id"`
	URL       *string    `json:"url" db:"url"`
	Platform  *string    `json:"platform" db:"platform"`
	Language  *string    `json:"language" db:"language"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

type broadcastURL struct {
	URL      *string `json:"url" db:"url" validate:"required"`
	Platform *string `json:"platform" db:"platform" validate:"required"`
	Language *string `json:"language" db:"language" validate:"required"`
}

type BroadcastURL interface {
	GetBroadcastURLByID(ctx *gin.Context)
	GetAllBroadcastURL(ctx *gin.Context)
	CreateNewBroadcastURL(ctx *gin.Context)
	UpdateBroadcastURLByID(ctx *gin.Context)
	DeleteBroadcastURLByID(ctx *gin.Context)
}

type BroadcastURLDB struct {
	db *pgxpool.Pool
}

func NewBroadcastURL(db *pgxpool.Pool) BroadcastURL {
	return &BroadcastURLDB{
		db,
	}
}

func (r *BroadcastURLDB) GetBroadcastURLByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getURLByID(r, ctx, id)

	if err != nil {
		if err.Error() == "not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *BroadcastURLDB) GetAllBroadcastURL(ctx *gin.Context) {
	skip := ctx.Query("skip")
	limit := ctx.Query("limit")

	if skip == "" {
		skip = "0"
	}

	if limit == "" {
		limit = "10"
	}

	// String conversion to int
	intSkip, err := strconv.Atoi(skip)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER", "success": false})
		return
	}

	// String conversion to int
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER", "success": false})
		return
	}

	u, err := getAllURL(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *BroadcastURLDB) CreateNewBroadcastURL(ctx *gin.Context) {
	s := broadcastURL{}
	if err := ctx.ShouldBindJSON(&s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	err := validator.New().Struct(s)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	if err := createNewURL(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Broadcast url!", "data": s, "success": true})
}

func (r *BroadcastURLDB) UpdateBroadcastURLByID(ctx *gin.Context) {
	u := broadcastURL{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateURLByID(r, ctx, u, id); err != nil {

		if err.Error() == "not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		if err.Error() == "invalid values" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Broadcast url updated successfully", "data": u, "success": true})
}

func (r *BroadcastURLDB) DeleteBroadcastURLByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteURLByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Broadcast url deleted successfully!", "success": true})
}

func getURLByID(r *BroadcastURLDB, ctx *gin.Context, id string) (broadcastURLResponse, error) {
	u := broadcastURLResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	url,
	platform,
	language,
	created_at,
	updated_at 
	from broadcast_url where id = $1`, id).Scan(
		&u.ID,
		&u.URL,
		&u.Platform,
		&u.Language,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return broadcastURLResponse{}, fmt.Errorf("not found")
		}
		return broadcastURLResponse{}, err
	}
	return u, nil
}

func getAllURL(r *BroadcastURLDB, ctx *gin.Context, skip int, limit int) (*[]broadcastURLResponse, error) {

	u := []broadcastURLResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	id,
	url,
	platform,
	language,
	created_at,
	updated_at 
	from broadcast_url LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d broadcastURLResponse
		err := rows.Scan(&d.ID, &d.URL, &d.Platform, &d.Language, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateURLByID(r *BroadcastURLDB, ctx *gin.Context, req broadcastURL, id string) error {

	toUpdate, toUpdateArgs := prepareURLUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE broadcast_url SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating broadcast url: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewURL(r *BroadcastURLDB, ctx *gin.Context, req broadcastURL) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO broadcast_url (
			url,
			platform,
			language)
		VALUES (
			$1,
			$2,
			$3)  `,
		*req.URL,
		*req.Platform,
		*req.Language)

	return err
}

func deleteURLByID(r *BroadcastURLDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from broadcast_url where id=$1", id)
	return err
}

func prepareURLUpdateQuery(req broadcastURL) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.URL != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("url=$%d", len(updateStrings)+1))
		args = append(args, *req.URL)
	}
	if req.Platform != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("platform=$%d", len(updateStrings)+1))
		args = append(args, *req.Platform)
	}
	if req.Language != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("language=$%d", len(updateStrings)+1))
		args = append(args, *req.Language)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
