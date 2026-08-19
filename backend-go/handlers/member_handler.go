package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ListProjectMembers lists every UserRole (admin + member) within one project —
// scoped strictly to projectID from context, unlike the global ListUsers.
func ListProjectMembers(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)

	query := database.DB.Table("user_roles").
		Select(`users.id, users.name, users.email, users.is_super_admin,
			user_roles.id as user_role_id, roles.name as role,
			projects.id as project_id, projects.name as project_name, users.created_at`).
		Joins("JOIN users ON users.id = user_roles.user_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Joins("JOIN projects ON projects.id = roles.project_id").
		Where("roles.project_id = ?", projectID)

	if term := searchTerm(c); term != "" {
		query = query.Where("LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?", term, term)
	}

	data := []UserListItem{}
	total, err := paginateScan(query.Order("users.created_at desc"), page, &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data member"})
		return
	}

	c.JSON(http.StatusOK, listResponse(data, total, page))
}

type AddMemberRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AddProjectMember creates a user and assigns them the "member" role in
// projectID (from context) — Admin can never assign "admin" via this route.
func AddProjectMember(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data member tidak lengkap atau tidak valid"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses password"})
		return
	}

	callerID := currentUserID(c)
	user, userRole, role, err := createUserWithRole(req.Name, req.Email, string(hashedPassword), callerID, projectID, "member")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menambahkan member: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"is_super_admin": user.IsSuperAdmin,
		"user_role_id":   userRole.ID,
		"role":           role.Name,
		"project_id":     projectID,
	})
}

// RemoveProjectMember revokes a UserRole, but only if it actually belongs to
// projectID from context — closes the cross-project tampering hole.
func RemoveProjectMember(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	userRoleID, ok := paramUint(c, "userRoleId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if !mustBelongToProject(userRoleID, projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member tidak ditemukan di project ini"})
		return
	}

	if err := database.DB.Where("user_role_id = ?", userRoleID).Delete(&models.MemberPermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus permission terkait"})
		return
	}

	if err := database.DB.Delete(&models.UserRole{}, userRoleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencabut member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member berhasil dicabut"})
}

// GetMyPermissions lets a member see their own granted permission codes for a
// project — used by the frontend sidebar to decide which nav items to show.
func GetMyPermissions(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	callerID := currentUserID(c)

	var userRole models.UserRole
	err := database.DB.
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.project_id = ? AND roles.name = ? AND user_roles.user_id = ?", projectID, "member", callerID).
		First(&userRole).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda bukan member project ini"})
		return
	}

	var granted []models.MemberPermission
	if err := database.DB.Where("user_role_id = ?", userRole.ID).Preload("Permission").Find(&granted).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data permission"})
		return
	}

	codes := make([]string, 0, len(granted))
	for _, g := range granted {
		codes = append(codes, g.Permission.Code)
	}

	c.JSON(http.StatusOK, gin.H{"data": codes})
}
