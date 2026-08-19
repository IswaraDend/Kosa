package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	defaultPerPage = 25
	maxPerPage     = 200
)

// Pagination is a resolved page request. PerPage 0 means "no limit" — the
// caller asked for the whole list.
type Pagination struct {
	Page    int
	PerPage int
}

// paginationFrom reads ?page= and ?per_page= off the query string.
//
// A request that names neither is answered unpaginated. That keeps every
// existing consumer working unchanged — the frontend hooks that grab
// res.data[0] to auto-select a project, ad-hoc curl calls, the import
// tooling — while any caller that does ask for a page gets one. per_page is
// clamped so a client cannot turn a page request into a full table scan.
func paginationFrom(c *gin.Context) Pagination {
	rawPage := c.Query("page")
	rawPerPage := c.Query("per_page")

	if rawPage == "" && rawPerPage == "" {
		return Pagination{Page: 1, PerPage: 0}
	}

	page, _ := strconv.Atoi(rawPage)
	if page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(rawPerPage)
	if err != nil || perPage <= 0 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return Pagination{Page: page, PerPage: perPage}
}

// applyPage counts the matching rows and returns a query narrowed to the
// requested slice.
//
// Both derived queries are taken as fresh Sessions: GORM mutates the statement
// it builds on, so counting and then fetching from the same *gorm.DB would let
// the count's changes (and the fetch's LIMIT) leak into each other. GORM drops
// ORDER BY for the count on its own when there is no GROUP BY.
func applyPage(query *gorm.DB, p Pagination) (*gorm.DB, int64, error) {
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageQuery := query.Session(&gorm.Session{})
	if p.PerPage > 0 {
		pageQuery = pageQuery.Limit(p.PerPage).Offset((p.Page - 1) * p.PerPage)
	}
	return pageQuery, total, nil
}

// paginate fills dest with one page of a model-backed query and reports how
// many rows matched in total.
func paginate(query *gorm.DB, p Pagination, dest interface{}) (int64, error) {
	pageQuery, total, err := applyPage(query, p)
	if err != nil {
		return 0, err
	}
	return total, pageQuery.Find(dest).Error
}

// paginateScan is paginate for queries built with Table()/Select() that
// project into a custom struct rather than a model.
func paginateScan(query *gorm.DB, p Pagination, dest interface{}) (int64, error) {
	pageQuery, total, err := applyPage(query, p)
	if err != nil {
		return 0, err
	}
	return total, pageQuery.Scan(dest).Error
}

// listResponse is the envelope every list endpoint returns.
//
// `total` is the number of rows matching the filters, NOT the number returned
// in `data` — that distinction is the whole point of paginating, and it is
// what lets a client render "halaman 2 dari 7" without a second request.
func listResponse(data interface{}, total int64, p Pagination) gin.H {
	perPage := p.PerPage
	if perPage <= 0 {
		perPage = int(total)
	}

	totalPages := 1
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
		if totalPages < 1 {
			totalPages = 1
		}
	}

	return gin.H{
		"data":        data,
		"total":       total,
		"page":        p.Page,
		"per_page":    perPage,
		"total_pages": totalPages,
	}
}
