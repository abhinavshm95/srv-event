package event

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

type eventItemResponse struct {
	ID        *int       `json:"id" db:"id"`
	EventID   *int       `json:"event_id" db:"event_id"`
	ItemID    *int       `json:"item_id" db:"item_id"`
	Deleted   *bool      `json:"deleted" db:"deleted"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

type eventItem struct {
	EventID *int  `json:"event_id" db:"event_id" validate:"required"`
	ItemID  *int  `json:"item_id" db:"item_id" validate:"required"`
	Deleted *bool `json:"deleted" db:"deleted"`
}

type EventItem interface {
	GetEventItemByID(ctx *gin.Context)
	GetAllEventItem(ctx *gin.Context)
	CreateNewEventItem(ctx *gin.Context)
	UpdateEventItemByID(ctx *gin.Context)
	DeleteEventItemByID(ctx *gin.Context)
}

type EventItemDB struct {
	db *pgxpool.Pool
}

func NewEventItem(db *pgxpool.Pool) EventItem {
	return &EventItemDB{
		db,
	}
}

func (r *EventItemDB) GetEventItemByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getEventItemByID(r, ctx, id)

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

func (r *EventItemDB) GetAllEventItem(ctx *gin.Context) {
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

	u, err := getAllEventItem(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *EventItemDB) CreateNewEventItem(ctx *gin.Context) {
	s := eventItem{}
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

	if err := createNewEventItem(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Event Item!", "data": s, "success": true})
}

func (r *EventItemDB) UpdateEventItemByID(ctx *gin.Context) {
	u := eventItem{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateEventItemByID(r, ctx, u, id); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Event Item updated successfully", "data": u, "success": true})
}

func (r *EventItemDB) DeleteEventItemByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteEventItemByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Event Item deleted successfully!", "success": true})
}

func getEventItemByID(r *EventItemDB, ctx *gin.Context, id string) (eventItemResponse, error) {
	u := eventItemResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	event_id,
	item_id,
	deleted,
	created_at,
	updated_at 
	from event_item where id = $1`, id).Scan(
		&u.ID,
		&u.EventID,
		&u.ItemID,
		&u.Deleted,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return eventItemResponse{}, fmt.Errorf("not found")
		}
		return eventItemResponse{}, err
	}
	return u, nil
}

func getAllEventItem(r *EventItemDB, ctx *gin.Context, skip int, limit int) (*[]eventItemResponse, error) {

	u := []eventItemResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	id,
	event_id,
	item_id,
	deleted,
	created_at,
	updated_at 
	from event_item LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d eventItemResponse
		err := rows.Scan(&d.ID, &d.EventID, &d.ItemID, &d.Deleted, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateEventItemByID(r *EventItemDB, ctx *gin.Context, req eventItem, id string) error {

	toUpdate, toUpdateArgs := prepareEventItemUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE event_item SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating Event Item: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewEventItem(r *EventItemDB, ctx *gin.Context, req eventItem) error {

	createString, numString, createQueryArgs := prepareEventItemCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO event_item (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating event item: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func deleteEventItemByID(r *EventItemDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from event_item where id=$1", id)
	return err
}

func prepareEventItemUpdateQuery(req eventItem) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.EventID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("event_id=$%d", len(updateStrings)+1))
		args = append(args, *req.EventID)
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

func prepareEventItemCreateQuery(req eventItem) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.EventID != nil {
		createStrings = append(createStrings, "event_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EventID)
	}
	if req.ItemID != nil {
		createStrings = append(createStrings, "item_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ItemID)
	}
	if req.Deleted != nil {
		createStrings = append(createStrings, "deleted")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Deleted)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
