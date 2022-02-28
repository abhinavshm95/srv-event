package item

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

type itemBroadcastURLResponse struct {
	ID             *int       `json:"id" db:"id"`
	ItemID         *int       `json:"item_id" db:"item_id"`
	BoradcastURLID *int       `json:"broadcast_url_id" db:"broadcast_url_id"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
}

type itemBroadcastURL struct {
	ItemID         *int `json:"item_id" db:"item_id" validate:"required"`
	BoradcastURLID *int `json:"broadcast_url_id" db:"broadcast_url_id" validate:"required"`
}

type ItemBroadcastURL interface {
	GetItemBroadcastURLByID(ctx *gin.Context)
	GetAllItemBroadcastURL(ctx *gin.Context)
	CreateNewItemBroadcastURL(ctx *gin.Context)
	UpdateItemBroadcastURLByID(ctx *gin.Context)
	DeleteItemBroadcastURLByID(ctx *gin.Context)
}

type ItemBroadcastURLDB struct {
	db *pgxpool.Pool
}

func NewItemBroadcastURL(db *pgxpool.Pool) ItemBroadcastURL {
	return &ItemBroadcastURLDB{
		db,
	}
}

func (r *ItemBroadcastURLDB) GetItemBroadcastURLByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getItemBroadcastURLByID(r, ctx, id)

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

func (r *ItemBroadcastURLDB) GetAllItemBroadcastURL(ctx *gin.Context) {
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

	u, err := getAllItemBroadcastURL(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *ItemBroadcastURLDB) CreateNewItemBroadcastURL(ctx *gin.Context) {
	s := itemBroadcastURL{}
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

	if err := createNewItemBroadcastURL(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Item BroadcastURL!", "data": s, "success": true})
}

func (r *ItemBroadcastURLDB) UpdateItemBroadcastURLByID(ctx *gin.Context) {
	u := itemBroadcastURL{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateItemBroadcastURLByID(r, ctx, u, id); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Item BroadcastURL updated successfully", "data": u, "success": true})
}

func (r *ItemBroadcastURLDB) DeleteItemBroadcastURLByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteItemBroadcastURLByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Item BroadcastURL deleted successfully!", "success": true})
}

func getItemBroadcastURLByID(r *ItemBroadcastURLDB, ctx *gin.Context, id string) (itemBroadcastURLResponse, error) {
	u := itemBroadcastURLResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	item_id,
	broadcast_url_id,
	created_at,
	updated_at 
	from item_broadcast_url where id = $1`, id).Scan(
		&u.ID,
		&u.ItemID,
		&u.BoradcastURLID,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return itemBroadcastURLResponse{}, fmt.Errorf("not found")
		}
		return itemBroadcastURLResponse{}, err
	}
	return u, nil
}

func getAllItemBroadcastURL(r *ItemBroadcastURLDB, ctx *gin.Context, skip int, limit int) (*[]itemBroadcastURLResponse, error) {

	u := []itemBroadcastURLResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	id,
	item_id,
	broadcast_url_id,
	created_at,
	updated_at 
	from item_broadcast_url LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d itemBroadcastURLResponse
		err := rows.Scan(&d.ID, &d.ItemID, &d.BoradcastURLID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateItemBroadcastURLByID(r *ItemBroadcastURLDB, ctx *gin.Context, req itemBroadcastURL, id string) error {

	toUpdate, toUpdateArgs := prepareItemBroadcastURLUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE item_broadcast_url SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating item broadcast url: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewItemBroadcastURL(r *ItemBroadcastURLDB, ctx *gin.Context, req itemBroadcastURL) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO item_broadcast_url (
			item_id,
			broadcast_url_id)
		VALUES (
			$1,
			$2)`,
		*req.ItemID,
		*req.BoradcastURLID)

	return err
}

func deleteItemBroadcastURLByID(r *ItemBroadcastURLDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from item_broadcast_url where id=$1", id)
	return err
}

func prepareItemBroadcastURLUpdateQuery(req itemBroadcastURL) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.BoradcastURLID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("broadcast_url_id=$%d", len(updateStrings)+1))
		args = append(args, *req.BoradcastURLID)
	}
	if req.ItemID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("item_id=$%d", len(updateStrings)+1))
		args = append(args, *req.ItemID)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
