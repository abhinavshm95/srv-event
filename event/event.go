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

type eventResponse struct {
	ID                   *int       `json:"id" db:"id"`
	RegistrationRequired *bool      `json:"registration_required" db:"registration_required"`
	RegistrationStatus   *string    `json:"registration_status" db:"registration_status"`
	Audience             *string    `json:"audience" db:"audience"`
	Slug                 *string    `json:"slug" db:"slug"`
	Name                 *string    `json:"name" db:"name"`
	Logo                 *string    `json:"logo,omitempty" db:"logo"`
	Content              *string    `json:"content,omitempty" db:"content"`
	Deleted              *bool      `json:"deleted" db:"deleted"`
	StartsOn             *time.Time `json:"starts_on" db:"starts_on"`
	EndsOn               *time.Time `json:"ends_on" db:"ends_on"`
	DateConfirmed        *bool      `json:"date_confirmed" db:"date_confirmed"`
	CreatedAt            *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at" db:"updated_at"`
}

type event struct {
	RegistrationRequired *bool      `json:"registration_required" db:"registration_required"`
	RegistrationStatus   *string    `json:"registration_status" db:"registration_status"`
	Audience             *string    `json:"audience" db:"audience"`
	Slug                 *string    `json:"slug" db:"slug" validate:"required"`
	Name                 *string    `json:"name" db:"name" validate:"required"`
	Logo                 *string    `json:"logo,omitempty" db:"logo"`
	Content              *string    `json:"content,omitempty" db:"content"`
	Deleted              *bool      `json:"deleted" db:"deleted"`
	StartsOn             *time.Time `json:"starts_on" db:"starts_on" validate:"required"`
	EndsOn               *time.Time `json:"ends_on" db:"ends_on" validate:"required"`
	DateConfirmed        *bool      `json:"date_confirmed" db:"date_confirmed"`
}

type Event interface {
	GetEventByID(ctx *gin.Context)
	GetAllEvent(ctx *gin.Context)
	CreateNewEvent(ctx *gin.Context)
	UpdateEventByID(ctx *gin.Context)
	DeleteEventByID(ctx *gin.Context)
	DeleteHardEventByID(ctx *gin.Context)
}

type EventDB struct {
	db *pgxpool.Pool
}

func NewEvent(db *pgxpool.Pool) Event {
	return &EventDB{
		db,
	}
}

func (r *EventDB) GetEventByID(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := getEventByID(r, ctx, id)

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

func (r *EventDB) GetAllEvent(ctx *gin.Context) {
	skip := ctx.Query("skip")
	limit := ctx.Query("limit")
	slug := ctx.Query("slug")

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

	fetchedEvents, err := getAllEvent(r, ctx, intSkip, intLimit, slug)

	// Manage if no event found
	if len(*fetchedEvents) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "no event found",
			"success": false,
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": fetchedEvents, "success": true})
}

func (r *EventDB) CreateNewEvent(ctx *gin.Context) {
	s := event{}
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

	if err := CreateEvent(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new Event!", "data": s, "success": true})
}

func (r *EventDB) UpdateEventByID(ctx *gin.Context) {
	u := event{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	id := ctx.Param("id")

	if err := updateEventByID(r, ctx, u, id); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Event updated successfully", "data": u, "success": true})
}

func (r *EventDB) DeleteEventByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteEventByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully!", "success": true})
}

func (r *EventDB) DeleteHardEventByID(ctx *gin.Context) {

	id := ctx.Param("id")

	if err := deleteHardEventByID(r, ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully!", "success": true})
}

func getEventByID(r *EventDB, ctx *gin.Context, id string) (eventResponse, error) {
	u := eventResponse{}
	if err := r.db.QueryRow(ctx, `select 
	id,
	registration_required,
	registration_status,
	audience,
	slug,
	name,
	logo,
	content,
	deleted,
	starts_on,
	ends_on,
	date_confirmed,
	created_at,
	updated_at 
	from event where id = $1`, id).Scan(
		&u.ID,
		&u.RegistrationRequired,
		&u.RegistrationStatus,
		&u.Audience,
		&u.Slug,
		&u.Name,
		&u.Logo,
		&u.Content,
		&u.Deleted,
		&u.StartsOn,
		&u.EndsOn,
		&u.DateConfirmed,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return eventResponse{}, fmt.Errorf("not found")
		}
		return eventResponse{}, err
	}
	return u, nil
}

func getAllEvent(r *EventDB, ctx *gin.Context, skip int, limit int, slug string) (*[]eventResponse, error) {

	whereQuery := buildAndGetWhereEventQuery(slug)

	u := []eventResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	id,
	registration_required,
	registration_status,
	audience,
	slug,
	name,
	logo,
	content,
	deleted,
	starts_on,
	ends_on,
	date_confirmed,
	created_at,
	updated_at 
	from event`+whereQuery+" LIMIT %d OFFSET %d", limit, skip))
	for rows.Next() {
		var d eventResponse
		err := rows.Scan(&d.ID, &d.RegistrationRequired, &d.RegistrationStatus, &d.Audience, &d.Slug, &d.Name, &d.Logo, &d.Content, &d.Deleted, &d.StartsOn, &d.EndsOn, &d.DateConfirmed, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func updateEventByID(r *EventDB, ctx *gin.Context, req event, id string) error {

	toUpdate, toUpdateArgs := prepareEventUpdateQuery(req)

	if len(toUpdateArgs) != 0 {
		updateRes, err := r.db.Exec(ctx, fmt.Sprintf(`UPDATE event SET %s WHERE id=%s`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating event: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func CreateEvent(r *EventDB, ctx *gin.Context, req event) error {

	createString, numString, createQueryArgs := prepareEventCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := r.db.Exec(ctx, fmt.Sprintf(`INSERT INTO event (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating event: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func deleteEventByID(r *EventDB, ctx context.Context, id string) error {

	eventQuery := "UPDATE event SET deleted = true WHERE id=" + id + ";"
	eventItemQuery := "UPDATE event_item SET deleted = true WHERE event_id=" + id + ";"
	eventPartQuery := "UPDATE event_participation_option SET deleted = true WHERE event_id=" + id + ";"
	eventStatusQuery := "UPDATE participation_status SET deleted = true WHERE event_id=" + id + ";"
	_, err := r.db.Exec(ctx, eventQuery+eventItemQuery+eventPartQuery+eventStatusQuery)
	return err
}

func deleteHardEventByID(r *EventDB, ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from event where id=$1", id)
	return err
}

func prepareEventUpdateQuery(req event) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.RegistrationRequired != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_required=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationRequired)
	}
	if req.RegistrationStatus != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("registration_status=$%d", len(updateStrings)+1))
		args = append(args, *req.RegistrationStatus)
	}
	if req.Audience != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("audience=$%d", len(updateStrings)+1))
		args = append(args, *req.Audience)
	}
	if req.Slug != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("slug=$%d", len(updateStrings)+1))
		args = append(args, *req.Slug)
	}
	if req.Name != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.Name)
	}
	if req.Logo != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("logo=$%d", len(updateStrings)+1))
		args = append(args, *req.Logo)
	}
	if req.Content != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("content=$%d", len(updateStrings)+1))
		args = append(args, *req.Content)
	}
	if req.Deleted != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("deleted=$%d", len(updateStrings)+1))
		args = append(args, *req.Deleted)
	}
	if req.StartsOn != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("starts_on=$%d", len(updateStrings)+1))
		args = append(args, *req.StartsOn)
	}
	if req.EndsOn != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("ends_on=$%d", len(updateStrings)+1))
		args = append(args, *req.EndsOn)
	}
	if req.DateConfirmed != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("date_confirmed=$%d", len(updateStrings)+1))
		args = append(args, *req.DateConfirmed)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareEventCreateQuery(req event) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.RegistrationRequired != nil {
		createStrings = append(createStrings, "registration_required")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationRequired)
	}
	if req.RegistrationStatus != nil {
		createStrings = append(createStrings, "registration_status")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RegistrationStatus)
	}
	if req.Audience != nil {
		createStrings = append(createStrings, "audience")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Audience)
	}
	if req.Slug != nil {
		createStrings = append(createStrings, "slug")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Slug)
	}
	if req.Name != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Name)
	}
	if req.Logo != nil {
		createStrings = append(createStrings, "logo")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Logo)
	}
	if req.Content != nil {
		createStrings = append(createStrings, "content")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Content)
	}
	if req.Deleted != nil {
		createStrings = append(createStrings, "deleted")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Deleted)
	}
	if req.StartsOn != nil {
		createStrings = append(createStrings, "starts_on")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.StartsOn)
	}
	if req.EndsOn != nil {
		createStrings = append(createStrings, "ends_on")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EndsOn)
	}
	if req.DateConfirmed != nil {
		createStrings = append(createStrings, "date_confirmed")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.DateConfirmed)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func buildAndGetWhereEventQuery(slug string) string {

	var whereString strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// WHERE query generation based on parameters
	if slug != "" {
		whereCondition.WriteString(fmt.Sprintf(" slug='%s'", slug))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String()
}
