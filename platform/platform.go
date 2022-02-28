package platform

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

type platformResponse struct {
	Name *string `json:"name" db:"name"`
}

type platform struct {
	Name *string `json:"Name" db:"Name" validate:"required"`
}

type Platform interface {
	GetPlatformByName(ctx *gin.Context)
	GetAllPlatform(ctx *gin.Context)
	CreateNewPlatform(ctx *gin.Context)
	UpdatePlatformByName(ctx *gin.Context)
	DeletePlatformByName(ctx *gin.Context)
}

type PlatformDB struct {
	db *pgxpool.Pool
}

func NewPlatform(db *pgxpool.Pool) Platform {
	return &PlatformDB{
		db,
	}
}

func (r *PlatformDB) GetPlatformByName(ctx *gin.Context) {
	name := ctx.Param("name")

	u, err := getPlatformByName(r, ctx, name)

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

func (r *PlatformDB) GetAllPlatform(ctx *gin.Context) {
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

	u, err := GetAllPlatform(r, ctx, intSkip, intLimit)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Fetched!", "data": u, "success": true})
}

func (r *PlatformDB) CreateNewPlatform(ctx *gin.Context) {
	s := platform{}
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

	if err := CreateNewPlatform(r, ctx, s); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Created new platform!", "data": s, "success": true})
}

func (r *PlatformDB) UpdatePlatformByName(ctx *gin.Context) {
	u := platform{}
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	name := ctx.Param("name")

	if err := UpdatePlatformByName(r, ctx, u, name); err != nil {

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
	ctx.JSON(http.StatusOK, gin.H{"message": "platform updated successfully", "data": u, "success": true})
}

func (r *PlatformDB) DeletePlatformByName(ctx *gin.Context) {

	name := ctx.Param("name")

	if err := DeletePlatformByName(r, ctx, name); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "platform deleted successfully!", "success": true})
}

func getPlatformByName(r *PlatformDB, ctx *gin.Context, name string) (platformResponse, error) {
	u := platformResponse{}
	if err := r.db.QueryRow(ctx, `select 
	name 
	from platform where name = $1`, name).Scan(
		&u.Name,
	); err != nil {
		if err == pgx.ErrNoRows {
			return platformResponse{}, fmt.Errorf("not found")
		}
		return platformResponse{}, err
	}
	return u, nil
}

func GetAllPlatform(r *PlatformDB, ctx *gin.Context, skip int, limit int) (*[]platformResponse, error) {

	u := []platformResponse{}
	rows, _ := r.db.Query(ctx, fmt.Sprintf(`select 
	name 
	from platform LIMIT %d OFFSET %d`, limit, skip))
	for rows.Next() {
		var d platformResponse
		err := rows.Scan(&d.Name)
		if err != nil {
			return &u, err
		}
		u = append(u, d)
	}
	return &u, rows.Err()
}

func UpdatePlatformByName(r *PlatformDB, ctx *gin.Context, req platform, name string) error {
	if req.Name != nil {
		updateRes, err := r.db.Exec(ctx, `UPDATE platform SET name=$1 WHERE name=$2`, req.Name, name)
		if err != nil {
			return fmt.Errorf("problem updating platform: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func CreateNewPlatform(r *PlatformDB, ctx *gin.Context, req platform) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO platform (
			name)
		VALUES (
			$1)  `,
		*req.Name)

	return err
}

func DeletePlatformByName(r *PlatformDB, ctx context.Context, name string) error {
	_, err := r.db.Exec(ctx, "delete from platform where name=$1", name)
	return err
}
