package handlers

import (
	"net/http"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

type ProjectResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	CreatedBy      uint      `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	AdminName      string    `json:"admin_name"`
	AdminCount     int64     `json:"admin_count"`
	MemberCount    int64     `json:"member_count"`
	WarehouseCount int64     `json:"warehouse_count"`
}

func buildProjectResponse(p models.Project) ProjectResponse {
	resp := ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		Status:      p.Status,
		CreatedBy:   p.AuthorID,
		CreatedAt:   p.CreatedAt,
	}

	database.DB.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.project_id = ? AND roles.name = ?", p.ID, "admin").
		Distinct("user_roles.user_id").
		Count(&resp.AdminCount)

	database.DB.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.project_id = ? AND roles.name = ?", p.ID, "member").
		Distinct("user_roles.user_id").
		Count(&resp.MemberCount)

	database.DB.Model(&models.Warehouse{}).Where("project_id = ?", p.ID).Count(&resp.WarehouseCount)

	var adminUser models.User
	err := database.DB.
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.project_id = ? AND roles.name = ?", p.ID, "admin").
		First(&adminUser).Error
	if err == nil {
		resp.AdminName = adminUser.Name
	}

	return resp
}

func ListProjects(c *gin.Context) {
	var projects []models.Project
	if err := database.DB.Order("created_at desc").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data project"})
		return
	}

	data := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		data = append(data, buildProjectResponse(p))
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "total": len(data)})
}

func GetProject(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, buildProjectResponse(project))
}

type ProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

func CreateProject(c *gin.Context) {
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan kode project wajib diisi"})
		return
	}

	project := models.Project{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      "aktif",
		AuthorID:    currentUserID(c),
	}

	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat project"})
		return
	}

	c.JSON(http.StatusCreated, buildProjectResponse(project))
}

func UpdateProject(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan kode project wajib diisi"})
		return
	}

	project.Name = req.Name
	project.Code = req.Code
	project.Description = req.Description

	if err := database.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui project"})
		return
	}

	c.JSON(http.StatusOK, buildProjectResponse(project))
}

type ProjectStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=aktif nonaktif"`
}

func ToggleProjectStatus(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	var req ProjectStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status harus 'aktif' atau 'nonaktif'"})
		return
	}

	project.Status = req.Status
	if err := database.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status project"})
		return
	}

	c.JSON(http.StatusOK, buildProjectResponse(project))
}

func listProjectsForUserRole(userID uint, roleName string) ([]models.Project, error) {
	var projects []models.Project
	err := database.DB.
		Joins("JOIN roles ON roles.project_id = projects.id").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("roles.name = ? AND user_roles.user_id = ?", roleName, userID).
		Distinct().
		Order("projects.created_at desc").
		Find(&projects).Error
	return projects, err
}

// ListMyProjectsAsAdmin lists only the projects where the caller holds an "admin" Role.
func ListMyProjectsAsAdmin(c *gin.Context) {
	projects, err := listProjectsForUserRole(currentUserID(c), "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data project"})
		return
	}
	data := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		data = append(data, buildProjectResponse(p))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "total": len(data)})
}

// ListMyProjectsAsMember lists only the projects where the caller holds a "member" Role.
func ListMyProjectsAsMember(c *gin.Context) {
	projects, err := listProjectsForUserRole(currentUserID(c), "member")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data project"})
		return
	}
	data := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		data = append(data, buildProjectResponse(p))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "total": len(data)})
}

// GetMyProject returns the project the middleware already validated the
// caller has access to (RequireProjectRole/RequirePermission set "projectID").
func GetMyProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, buildProjectResponse(project))
}

func DeleteProject(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	var warehouseCount, itemCount, transactionCount, roleCount int64
	database.DB.Model(&models.Warehouse{}).Where("project_id = ?", id).Count(&warehouseCount)
	database.DB.Model(&models.Item{}).Where("project_id = ?", id).Count(&itemCount)
	database.DB.Model(&models.Transaction{}).Where("project_id = ?", id).Count(&transactionCount)
	database.DB.Model(&models.Role{}).Where("project_id = ?", id).Count(&roleCount)

	if warehouseCount > 0 || itemCount > 0 || transactionCount > 0 || roleCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Project masih memiliki data terkait (gudang/item/transaksi/role), nonaktifkan saja"})
		return
	}

	if err := database.DB.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project berhasil dihapus"})
}
