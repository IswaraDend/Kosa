package handlers

import (
	"strconv"
	"strings"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func currentUserID(c *gin.Context) uint {
	v, _ := c.Get("userID")
	id, _ := v.(uint)
	return id
}

func paramID(c *gin.Context) (uint, bool) {
	return paramUint(c, "id")
}

func paramUint(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

func queryUintPtr(c *gin.Context, key string) *uint {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	val := uint(v)
	return &val
}

// conflictError signals a business-rule 409 (as opposed to an unexpected DB error)
// from a core function, so thin handlers can map it to the right HTTP status.
type conflictError struct {
	message string
}

func (e *conflictError) Error() string {
	return e.message
}

func errConflict(message string) error {
	return &conflictError{message: message}
}

// applyCreatedAtRange applies a from/to filter on a timestamp column.
//
// A bare "YYYY-MM-DD" upper bound parses as that day's midnight, so a plain
// `<= to` silently drops every record made later that same day — picking
// today as the end date would return nothing from today. When `to` carries no
// time component we therefore switch to an exclusive `< to+1day`, which is
// what choosing an end date actually means.
func applyCreatedAtRange(query *gorm.DB, column, from, to string) *gorm.DB {
	if from != "" {
		query = query.Where(column+" >= ?", from)
	}
	if to != "" {
		if day, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where(column+" < ?", day.AddDate(0, 0, 1))
		} else {
			query = query.Where(column+" <= ?", to)
		}
	}
	return query
}

// searchTerm returns the ?q= filter as a LIKE pattern, lower-cased, or "" when
// no search was requested.
//
// Search has to run in SQL once a list is paginated: filtering the page the
// server already sliced would only ever search 25 rows, which looks like data
// silently going missing.
func searchTerm(c *gin.Context) string {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		return ""
	}
	return "%" + strings.ToLower(q) + "%"
}

// isDuplicateKeyError reports whether err is a unique-constraint violation, so
// a caller that generates a value optimistically can tell "someone committed
// the same value first, retry" apart from a genuine failure. Covers Postgres
// (SQLSTATE 23505) and SQLite, both of which are wired into this module.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "23505") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint")
}

// mustBelongToProject checks a UserRole row belongs to the given project
// (via its Role.ProjectID), preventing cross-project tampering by ID guessing.
func mustBelongToProject(userRoleID, projectID uint) bool {
	var count int64
	database.DB.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.id = ? AND roles.project_id = ?", userRoleID, projectID).
		Count(&count)
	return count > 0
}
