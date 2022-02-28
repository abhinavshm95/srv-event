package partstatus

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

type participationStatusResponse struct {
	ID                  *int       `json:"id" db:"id"`
	ParticipationOption *string    `json:"participation_option" db:"participation_option"`
	ParticipantID       *int       `json:"participant_id" db:"participant_id"`
	EventID             *int       `json:"event_id" db:"event_id"`
	Confirmed           *bool      `json:"confirmed" db:"confirmed"`
	RegistrationDate    *time.Time `json:"registration_date" db:"registration_date"`
	Deleted             *bool      `json:"deleted" db:"deleted"`
	CreatedAt           *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at" db:"updated_at"`
}

type participationStatus struct {
	ParticipationOption *string    `json:"participation_option" db:"participation_option" validate:"required"`
	ParticipantID       *int       `json:"participant_id" db:"participant_id" validate:"required"`
	EventID             *int       `json:"event_id" db:"event_id" validate:"required"`
	Confirmed           *bool      `json:"confirmed" db:"confirmed"`
	RegistrationDate    *time.Time `json:"registration_date" db:"registration_date" validate:"required"`
	Deleted             *bool      `json:"deleted" db:"deleted"`
}

type ParticipationStatus interface {
	GetParticipationStatusByID(ctx *gin.Context)
	GetAllParticipationStatus(ctx *gin.Context)
	CreateNewParticipationStatus(ctx *gin.Context)
	UpdateParticipationStatusByID(ctx *gin.Context)
	DeleteParticipationStatusByID(ctx *gin.Context)
}

type ParticipationStatusDB struct {
	db *pgxpool.Pool
}

func NewParticipationStatus(db *pgxpool.Pool) ParticipationStatus {
	return &ParticipationStatusDB{
		db,
	}
}

func (r *ParticipationStatusDB) GetParticipationStatusByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getParticipationStatusByID(r, ctx, id)

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

func (r *ParticipationStatusDB) GetAllParticipationStatus(ctx *gin.Context) {
	skip := ctx.Query("skip")
	limit := ctx.Query("limit")
	eventID := ctx.Query("eventid")

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

	u, err := getAllParticipationStatus(r, ctx, intSkip, intLimit, eventID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *ParticipationStatusDB) CreateNewParticipationStatus(ctx *gin.Context) {
	s := participationStatus{}
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

	if err := createNewParticipationStatus(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Participation Status!", "data": s, "success": true})
}

func (r *ParticipationStatusDB) UpdateParticipationStatusByID(ctx *gin.Context) {
	u := participationStatus{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateParticipationStatusByID(r, ctx, u, id); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Participation Status updated successfully", "data": u, "success": true})
}

func (r *ParticipationStatusDB) DeleteParticipationStatusByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteParticipationStatusByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Participation Status deleted successfully!", "success": true})
}

func getParticipationStatusByID(r *ParticipationStatusDB, ctx *gin.Context, id string) (participationStatusResponse, error) {
	u := participationStatusResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	participation_option,
	participant_id,
	event_id,
	confirmed,
	registration_date,
	deleted,
	created_at,
	updated_at 
	from participation_status where id = $1`, id).Scan(
		&u.ID,
		&u.ParticipationOption,
		&u.ParticipantID,
		&u.EventID,
		&u.Confirmed,
		&u.RegistrationDate,
		&u.Deleted,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return participationStatusResponse{}, fmt.Errorf("not found")
		}
		return participationStatusResponse{}, err
	}
	return u, nil
}

func getAllParticipationStatus(r *ParticipationStatusDB, ctx *gin.Context, skip int, limit int, eventID string) (*[]participationStatusResponse, error) {

	u := []participationStatusResponse{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereQuery(eventID)

	rows, err := r.db.Query(ctx, `select 
	id,
	participation_option,
	participant_id,
	event_id,
	confirmed,
	registration_date,
	deleted,
	created_at,
	updated_at 
	from participation_status`+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", limit, skip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return &u, err
	}
	defer rows.Close()
	for rows.Next() {
		var d participationStatusResponse
		err := rows.Scan(&d.ID, &d.ParticipationOption, &d.ParticipantID, &d.EventID, &d.Confirmed, &d.RegistrationDate, &d.Deleted, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateParticipationStatusByID(r *ParticipationStatusDB, ctx *gin.Context, req participationStatus, id string) error {

	toUpdate, toUpdateArgs := prepareParticipationStatusUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE participation_status SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating Participation Status: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func createNewParticipationStatus(r *ParticipationStatusDB, ctx *gin.Context, req participationStatus) error {

	createString, numString, createQueryArgs := prepareParticipationStatusCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO participation_status (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating participation status: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func deleteParticipationStatusByID(r *ParticipationStatusDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from participation_status where id=$1", id)
	return err
}

func prepareParticipationStatusUpdateQuery(req participationStatus) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.ParticipationOption != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participation_option=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipationOption)
	}
	if req.ParticipantID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("participant_id=$%d", len(updateStrings)+1))
		args = append(args, *req.ParticipantID)
	}
	if req.EventID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("event_id=$%d", len(updateStrings)+1))
		args = append(args, *req.EventID)
	}
	if req.Confirmed != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("confirmed=$%d", len(updateStrings)+1))
		args = append(args, *req.Confirmed)
	}
	if req.RegistrationDate != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_date=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationDate)
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

func prepareParticipationStatusCreateQuery(req participationStatus) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.ParticipationOption != nil {
		createStrings = append(createStrings, "participation_option")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ParticipationOption)
	}
	if req.ParticipantID != nil {
		createStrings = append(createStrings, "participant_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.ParticipantID)
	}
	if req.EventID != nil {
		createStrings = append(createStrings, "event_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EventID)
	}
	if req.Confirmed != nil {
		createStrings = append(createStrings, "confirmed")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Confirmed)
	}
	if req.RegistrationDate != nil {
		createStrings = append(createStrings, "registration_date")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationDate)
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

func buildAndGetWhereQuery(eventID string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// WHERE query generation based on parameters
	if eventID != "" {
		whereCondition.WriteString(fmt.Sprintf(" event_id=%s", eventID))
	}

	orderBy.WriteString(fmt.Sprintf(" ORDER BY created_at %s", "asc"))

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}
