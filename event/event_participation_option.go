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

type eventPartOptionResponse struct {
	ID                  *int       `json:"id" db:"id"`
	EventID             *int       `json:"event_id" db:"event_id"`
	ParticipationOption *string    `json:"participation_option" db:"participation_option"`
	Deleted             *bool      `json:"deleted" db:"deleted"`
	CreatedAt           *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at" db:"updated_at"`
}

type eventPartOption struct {
	EventID             *int    `json:"event_id" db:"event_id" validate:"required"`
	ParticipationOption *string `json:"participation_option" db:"participation_option" validate:"required"`
	Deleted             *bool   `json:"deleted" db:"deleted"`
}

type EventPartOption interface {
	GetEventPartOptionByID(ctx *gin.Context)
	GetAllEventPartOption(ctx *gin.Context)
	CreateNewEventPartOption(ctx *gin.Context)
	UpdateEventPartOptionByID(ctx *gin.Context)
	DeleteEventPartOptionByID(ctx *gin.Context)
}

type EventPartOptionDB struct {
	db *pgxpool.Pool
}

func NewEventPartOption(db *pgxpool.Pool) EventPartOption {
	return &EventPartOptionDB{
		db,
	}
}

func (r *EventPartOptionDB) GetEventPartOptionByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getEventPartOptionByID(r, ctx, id)

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

func (r *EventPartOptionDB) GetAllEventPartOption(ctx *gin.Context) {
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

	u, err := getAllEventPartOption(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *EventPartOptionDB) CreateNewEventPartOption(ctx *gin.Context) {
	s := eventPartOption{}
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

	if err := createNewEventPartOption(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Event Participation Option!", "data": s, "success": true})
}

func (r *EventPartOptionDB) UpdateEventPartOptionByID(ctx *gin.Context) {
	u := eventPartOption{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateEventPartOptionByID(r, ctx, u, id); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Event Participation Option updated successfully", "data": u, "success": true})
}

func (r *EventPartOptionDB) DeleteEventPartOptionByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteEventPartOptionByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Event Participation Option deleted successfully!", "success": true})
}

func getEventPartOptionByID(r *EventPartOptionDB, ctx *gin.Context, id string) (eventPartOptionResponse, error) {
	u := eventPartOptionResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	event_id,
	participation_option,
	deleted,
	created_at,
	updated_at 
	from event_participation_option where id = $1`, id).Scan(
		&u.ID,
		&u.EventID,
		&u.ParticipationOption,
		&u.Deleted,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return eventPartOptionResponse{}, fmt.Errorf("not found")
		}
		return eventPartOptionResponse{}, err
	}
	return u, nil
}

func getAllEventPartOption(r *EventPartOptionDB, ctx *gin.Context, skip int, limit int) (*[]eventPartOptionResponse, error) {

	u := []eventPartOptionResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	id,
	event_id,
	participation_option,
	deleted,
	created_at,
	updated_at 
	from event_participation_option LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d eventPartOptionResponse
		err := rows.Scan(&d.ID, &d.EventID, &d.ParticipationOption, &d.Deleted, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateEventPartOptionByID(r *EventPartOptionDB, ctx *gin.Context, req eventPartOption, id string) error {

	toUpdate, toUpdateArgs := prepareEventPartOptionUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE event_participation_option SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating Event Participation Option: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewEventPartOption(r *EventPartOptionDB, ctx *gin.Context, req eventPartOption) error {

	createString, numString, createQueryArgs := prepareEventPartOptionCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO event_participation_option (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating participation status: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func deleteEventPartOptionByID(r *EventPartOptionDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from event_participation_option where id=$1", id)
	return err
}

func prepareEventPartOptionUpdateQuery(req eventPartOption) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.EventID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("event_id=$%d", len(updateStrings)+1))
		args = append(args, *req.EventID)
	}
	if req.ParticipationOption != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participation_option=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipationOption)
	}
	if req.Deleted != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("deleted=$%d", len(updateStrings)+1))
		args = append(args, *req.Deleted)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareEventPartOptionCreateQuery(req eventPartOption) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.EventID != nil {
		createStrings = append(createStrings, "event_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EventID)
	}
	if req.ParticipationOption != nil {
		createStrings = append(createStrings, "participation_option")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ParticipationOption)
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
