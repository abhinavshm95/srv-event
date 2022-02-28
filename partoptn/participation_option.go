package partoptn

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type partOptionResponse struct {
	Name *string `json:"name" db:"name"`
}

type partOption struct {
	Name *string `json:"Name" db:"Name" validate:"required"`
}

type ParticipationOption interface {
	GetParticipationOptionByName(ctx *gin.Context)
	GetAllParticipationOption(ctx *gin.Context)
	CreateNewParticipationOption(ctx *gin.Context)
	UpdateParticipationOptionByName(ctx *gin.Context)
	DeleteParticipationOptionByName(ctx *gin.Context)
}

type ParticipationOptionDB struct {
	db *pgxpool.Pool
}

func NewParticipationOption(db *pgxpool.Pool) ParticipationOption {
	return &ParticipationOptionDB{
		db,
	}
}

func (r *ParticipationOptionDB) GetParticipationOptionByName(ctx *gin.Context) {
	name := ctx.Param("name")

	u, err := getPartOptionByName(r, ctx, name)

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

func (r *ParticipationOptionDB) GetAllParticipationOption(ctx *gin.Context) {
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

	u, err := GetAllPartOption(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *ParticipationOptionDB) CreateNewParticipationOption(ctx *gin.Context) {
	s := partOption{}
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

	if err := CreateNewPartOption(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new participation option!", "data": s, "success": true})
}

func (r *ParticipationOptionDB) UpdateParticipationOptionByName(ctx *gin.Context) {
	u := partOption{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	name := ctx.Param("name")

	if err := UpdatePartOptionByName(r, ctx, u, name); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "Participation option updated successfully", "data": u, "success": true})
}

func (r *ParticipationOptionDB) DeleteParticipationOptionByName(ctx *gin.Context) {

	name := ctx.Param("name")

	if err := DeletePartOptionByName(r, ctx, name); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Participation option deleted successfully!", "success": true})
}

func getPartOptionByName(r *ParticipationOptionDB, ctx *gin.Context, name string) (partOptionResponse, error) {
	u := partOptionResponse{}
	if err := r.db.QueryRow(ctx, `select 
	name 
	from participation_option where name = $1`, name).Scan(
		&u.Name,
	); err != nil {
		if err == pgx.ErrNoRows {
			return partOptionResponse{}, fmt.Errorf("not found")
		}
		return partOptionResponse{}, err
	}
	return u, nil
}

func GetAllPartOption(r *ParticipationOptionDB, ctx *gin.Context, skip int, limit int) (*[]partOptionResponse, error) {

	u := []partOptionResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	name 
	from participation_option LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d partOptionResponse
		err := rows.Scan(&d.Name)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func UpdatePartOptionByName(r *ParticipationOptionDB, ctx *gin.Context, req partOption, name string) error {
	if req.Name != nil {
		updateRes, err := r.db.Exec(ctx, `UPDATE participation_option SET name=$1 WHERE name=$2`, req.Name, name)
		if err != nil {
			return fmt.Errorf("problem updating participation_option: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func CreateNewPartOption(r *ParticipationOptionDB, ctx *gin.Context, req partOption) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO participation_option (
			name)
		VALUES (
			$1)  `,
		*req.Name)

	return err
}

func DeletePartOptionByName(r *ParticipationOptionDB, ctx context.Context, name string) error {
	_, err := r.db.Exec(ctx, "delete from participation_option where name=$1", name)
	return err
}
