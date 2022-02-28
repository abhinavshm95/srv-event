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

type itemResponse struct {
	ID               *int       `json:"id" db:"id"`
	StartDate        *time.Time `json:"start_date" db:"start_date"`
	Duration         *int       `json:"duration" db:"duration"`
	Name             *string    `json:"name" db:"name"`
	Content          *string    `json:"content,omitempty" db:"content"`
	OriginalLanguage *string    `json:"original_language" db:"original_language"`
	Translated       *bool      `json:"translated" db:"translated"`
	CreatedAt        *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at" db:"updated_at"`
}

type item struct {
	StartDate        *time.Time `json:"start_date" db:"start_date" validate:"required"`
	Duration         *int       `json:"duration" db:"duration" validate:"required"`
	Name             *string    `json:"name" db:"name" validate:"required"`
	Content          *string    `json:"content,omitempty" db:"content"`
	OriginalLanguage *string    `json:"original_language" db:"original_language" validate:"required"`
	Translated       *bool      `json:"translated" db:"translated" validate:"required"`
}

type Item interface {
	GetItemByID(ctx *gin.Context)
	GetAllItem(ctx *gin.Context)
	CreateNewItem(ctx *gin.Context)
	UpdateItemByID(ctx *gin.Context)
	DeleteItemByID(ctx *gin.Context)
}

type ItemDB struct {
	db *pgxpool.Pool
}

func NewItem(db *pgxpool.Pool) Item {
	return &ItemDB{
		db,
	}
}

func (r *ItemDB) GetItemByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getItemByID(r, ctx, id)

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

func (r *ItemDB) GetAllItem(ctx *gin.Context) {
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

	u, err := getAllItem(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *ItemDB) CreateNewItem(ctx *gin.Context) {
	s := item{}
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

	if err := createNewItem(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Item!", "data": s, "success": true})
}

func (r *ItemDB) UpdateItemByID(ctx *gin.Context) {
	u := item{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateItemByID(r, ctx, u, id); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Item updated successfully", "data": u, "success": true})
}

func (r *ItemDB) DeleteItemByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteItemByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully!", "success": true})
}

func getItemByID(r *ItemDB, ctx *gin.Context, id string) (itemResponse, error) {
	u := itemResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	start_date,
	duration,
	name,
	original_language,
	translated,
	created_at,
	updated_at 
	from item where id = $1`, id).Scan(
		&u.ID,
		&u.StartDate,
		&u.Duration,
		&u.Name,
		&u.OriginalLanguage,
		&u.Translated,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return itemResponse{}, fmt.Errorf("not found")
		}
		return itemResponse{}, err
	}
	return u, nil
}

func getAllItem(r *ItemDB, ctx *gin.Context, skip int, limit int) (*[]itemResponse, error) {

	u := []itemResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	id,
	start_date,
	duration,
	name,
	original_language,
	translated,
	created_at,
	updated_at 
	from item LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d itemResponse
		err := rows.Scan(&d.ID, &d.StartDate, &d.Duration, &d.Name, &d.OriginalLanguage, &d.Translated, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateItemByID(r *ItemDB, ctx *gin.Context, req item, id string) error {

	toUpdate, toUpdateArgs := prepareItemUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE item SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating item: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewItem(r *ItemDB, ctx *gin.Context, req item) error {
	createString, numString, createQueryArgs := prepareItemCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO item (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating item: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func deleteItemByID(r *ItemDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from item where id=$1", id)
	return err
}

func prepareItemUpdateQuery(req item) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.StartDate != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("start_date=$%d", len(updateStrings)+1))
		args = append(args, *req.StartDate)
	}
	if req.Duration != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("duration=$%d", len(updateStrings)+1))
		args = append(args, *req.Duration)
	}
	if req.Name != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.Name)
	}
	if req.OriginalLanguage != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("original_language=$%d", len(updateStrings)+1))
		args = append(args, *req.OriginalLanguage)
	}
	if req.Translated != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("translated=$%d", len(updateStrings)+1))
		args = append(args, *req.Translated)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareItemCreateQuery(req item) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.StartDate != nil {
		createStrings = append(createStrings, "start_date")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.StartDate)
	}
	if req.Duration != nil {
		createStrings = append(createStrings, "duration")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Duration)
	}
	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}
	if req.Content != nil {
		createStrings = append(createStrings, "content")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Content)
	}
	if req.OriginalLanguage != nil {
		createStrings = append(createStrings, "original_language")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.OriginalLanguage)
	}
	if req.Translated != nil {
		createStrings = append(createStrings, "translated")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Translated)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
