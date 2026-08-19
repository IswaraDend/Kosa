package handlers

import (
	"net/http"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	Modules        []string  `json:"modules"`
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

	resp.Modules = getProjectModules(p.ID)

	return resp
}

func getProjectModules(projectID uint) []string {
	var rows []models.ProjectModule
	database.DB.Where("project_id = ?", projectID).Find(&rows)
	modules := make([]string, 0, len(rows))
	for _, r := range rows {
		modules = append(modules, r.Module)
	}
	return modules
}

// pruneMemberPermissions deletes every MemberPermission in the project whose
// permission belongs to a module that is not in `enabled`.
//
// Without this, disabling a module leaves grants behind that no screen shows
// (the permission catalogue only lists enabled modules) and no request honours
// (RequireModule rejects them anyway) — so re-enabling the module later would
// silently restore access somebody thought they had revoked.
func pruneMemberPermissions(tx *gorm.DB, projectID uint, enabled []string) error {
	userRoleIDs := tx.Model(&models.UserRole{}).
		Select("user_roles.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.project_id = ?", projectID)

	query := tx.Where("user_role_id IN (?)", userRoleIDs)

	// An empty module set means every grant in the project is stale, so the
	// permission filter is skipped entirely rather than built as "NOT IN ()".
	if len(enabled) > 0 {
		stalePermissionIDs := tx.Model(&models.Permission{}).
			Select("id").Where("module NOT IN ?", enabled)
		query = query.Where("permission_id IN (?)", stalePermissionIDs)
	}

	return query.Delete(&models.MemberPermission{}).Error
}

// setProjectModules replaces the full set of enabled modules for a project:
// expands requested modules to include their dependencies (see
// models.ExpandModules), then deletes and re-inserts the ProjectModule rows
// in one DB transaction so a partial write can never leave a stale mix.
// Member grants for now-disabled modules are pruned in the same transaction.
func setProjectModules(projectID uint, requested []string) error {
	valid := make([]string, 0, len(requested))
	for _, m := range requested {
		if models.IsValidModule(m) {
			valid = append(valid, m)
		}
	}
	expanded := models.ExpandModules(valid)

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", projectID).Delete(&models.ProjectModule{}).Error; err != nil {
			return err
		}
		for _, m := range expanded {
			if err := tx.Create(&models.ProjectModule{ProjectID: projectID, Module: m}).Error; err != nil {
				return err
			}
		}
		return pruneMemberPermissions(tx, projectID, expanded)
	})
}

func ListProjects(c *gin.Context) {
	page := paginationFrom(c)
	query := database.DB.Model(&models.Project{})
	if term := searchTerm(c); term != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", term, term)
	}

	var projects []models.Project
	total, err := paginate(query.Order("created_at desc"), page, &projects)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data project"})
		return
	}

	data := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		data = append(data, buildProjectResponse(p))
	}

	c.JSON(http.StatusOK, listResponse(data, total, page))
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
	Name        string   `json:"name" binding:"required"`
	Code        string   `json:"code" binding:"required"`
	Description string   `json:"description"`
	Modules     []string `json:"modules"`
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

	if err := setProjectModules(project.ID, req.Modules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan modul project"})
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

	if err := setProjectModules(project.ID, req.Modules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan modul project"})
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

// listProjectsForUserRole is deliberately NOT paginated: it returns only the
// projects one user belongs to, which is bounded by their own memberships, and
// the frontend hooks that auto-select a project (useProjectAutoSelect,
// useMemberPermissions) need the whole set to decide which one is current.
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
