package services

import (
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// @Summary Get domain by id
// @Description Get domain by id
// @Description user must have read permission to domain or be admin
// @Tags domain
// @Param id path string true "domain id"
// @Produce json
// @Success 200 {object} mw.Domain{data=models.Domain}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id} [get]
func DomainGet(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user admin
	isAdmin := (c.Locals("role").(models.UserRole) >= models.Admin)

	// check if user has permission
	hasPermission := checkUserDomainPermission(uId, qId, models.ReadOnly)

	// check if user can access domain
	if !(isAdmin || hasPermission) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// get domain info
	var domain models.Domain
	if err := db.DB.Where("id = ?", qId).First(&domain).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "domain not found",
			Data:   qId,
		})
	}

	// admin has no permission to privacy domain
	if !hasPermission && isAdmin && domain.Privacy == true {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied, privacy domain",
			Data:   qId,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   domain,
	})
}

// @Summary Create domain
// @Description Create domain
// @Description user must have contributor role or higher
// @Tags domain
// @Accept json
// @Produce json
// @Param domain body mw.DomainInfoUpdate true "domain info"
// @Success 200 {object} mw.Domain{data=models.Domain}
// @Failure 400 {object} mw.Domain{data=mw.DomainInfoUpdate}
// @Failure 403 {object} mw.Domain{data=string}
// @Failure 500 {object} mw.Domain{data=string}
// @Router /api/v1/domain [post]
func DomainCreate(c *fiber.Ctx) error {
	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user role level is high enough to create domain
	if !(c.Locals("role").(models.UserRole) >= models.Contributor) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   uId,
		})
	}

	// parse request body
	var domain mw.DomainInfoUpdate
	if err := c.BodyParser(&domain); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: "invalid request body",
			Data:   domain,
		})
	}

	// check if query is valid
	if domain.Name == nil || domain.ApiId == nil || domain.ApiSecret == nil || domain.Vendor == nil || domain.ICPReg == nil {
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: "invalid request body",
			Data:   domain,
		})
	}

	// add domain and grant user owner rights to domain with transaction
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		// add domain
		d := models.Domain{
			Name:      *domain.Name,
			ApiId:     *domain.ApiId,
			ApiSecret: *domain.ApiSecret,
			Vendor:    *domain.Vendor,
			ICPReg:    *domain.ICPReg,
			Privacy:   *domain.Privacy,
		}
		if err := tx.Create(&d).Error; err != nil {
			return err
		}

		// grant user owner rights to domain
		ud := models.UserDomain{
			UserId:   uint(uId),
			DomainId: d.ID,
			Role:     models.Owner,
		}
		if err := tx.Create(&ud).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		logrus.Errorf("create domain error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   domain,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(mw.Domain{
		Status: fiber.StatusCreated,
		Data:   domain,
	})
}

// @Summary Update domain
// @Description Update domain
// @Description user must have manager role to domain or be admin
// @Description **ICP_reg param can't be updated**
// @Tags domain
// @Accept json
// @Produce json
// @Param id path string true "domain id"
// @Param domain body mw.DomainInfoUpdate true "domain info"
// @Success 200 {object} mw.Domain{data=models.Domain}
// @Failure 400 {object} mw.Domain{data=mw.DomainInfoUpdate}
// @Failure 403 {object} mw.Domain{data=string}
// @Failure 500 {object} mw.Domain{data=string}
// @Router /api/v1/domain/{id} [put]
func DomainUpdate(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user role level
	isAdmin := (c.Locals("role").(models.UserRole) >= models.Admin)
	updatePermitted := checkUserDomainPermission(uId, qId, models.Manager)
	if !(isAdmin || updatePermitted) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// parse request body
	var domain mw.DomainInfoUpdate
	if err := c.BodyParser(&domain); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: "invalid request body",
			Data:   domain,
		})
	}

	var d models.Domain
	if err := db.DB.Where("id = ?", qId).First(&d).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "domain not found",
			Data:   qId,
		})
	}
	if !updatePermitted && isAdmin && d.Privacy == true {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied, privacy domain",
			Data:   qId,
		})
	}

	// update domain info
	d.Name = utils.IfThenPtr(domain.Name, d.Name)
	d.ApiId = utils.IfThenPtr(domain.ApiId, d.ApiId)
	d.ApiSecret = utils.IfThenPtr(domain.ApiSecret, d.ApiSecret)
	d.Vendor = utils.IfThenPtr(domain.Vendor, d.Vendor)
	// d.ICPReg = utils.IfThenPtr(domain.ICPReg, d.ICPReg)
	if err := db.DB.Save(&d).Error; err != nil {
		logrus.Errorf("update domain error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   domain,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   domain,
	})
}

// @Summary Delete domain
// @Description Delete domain
// @Description user must have owner role to domain or be admin
// @Tags domain
// @Accept json
// @Produce json
// @Param id path string true "domain id"
// @Success 200 {object} mw.Domain{data=string}
// @Failure 403 {object} mw.Domain{data=string}
// @Failure 404 {object} mw.Domain{data=string}
// @Failure 500 {object} mw.Domain{data=string}
// @Router /api/v1/domain/{id} [delete]
func DomainDelete(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user role level
	isAdmin := (c.Locals("role").(models.UserRole) >= models.Admin)
	deletePermitted := checkUserDomainPermission(uId, qId, models.Owner)
	if !(isAdmin || deletePermitted) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// get domain info
	var domain models.Domain
	if err := db.DB.Where("id = ?", qId).First(&domain).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "domain not found",
			Data:   qId,
		})
	}

	if !deletePermitted && isAdmin && domain.Privacy == true {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied, privacy domain",
			Data:   qId,
		})
	}

	// delete domain
	if err := db.DB.Delete(&domain).Error; err != nil {
		logrus.Errorf("delete domain error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   domain,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   domain,
	})
}

// @Summary List domains
// @Description List domains
// @Description user can list all domains if user role level is admin
// @Description user can list domains which user has read access if user role level is not admin
// @Tags domain
// @Produce json
// @Success 200 {object} mw.Domain{data=[]models.Domain}
// @Failure 500 {object} mw.Domain{data=string}
// @Router /api/v1/domain [get]
func DomainList(c *fiber.Ctx) error {
	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user role level
	var domains []models.Domain
	if c.Locals("role").(models.UserRole) >= models.Admin {
		// get all domains
		if err := db.DB.Find(&domains).Error; err != nil {
			logrus.Errorf("get all domains error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   domains,
			})
		}

		// admin shouldn't see the domain with privacy
		domainWithNoPrivacy := filterDomainWithNoPrivacy(domains)

		return c.Status(fiber.StatusOK).JSON(mw.Domain{
			Status: fiber.StatusOK,
			Data:   domainWithNoPrivacy,
		})
	}

	// get domains which user can access with userDomain join
	if err := db.DB.Model(&models.User{Model: gorm.Model{ID: uId}}).Association("Domains").Find(&domains); err != nil {
		logrus.Errorf("get user domains error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   domains,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   domains,
	})
}

func filterDomainWithNoPrivacy(domains []models.Domain) []models.Domain {
	var noPrivacyDomains []models.Domain
	for i := range domains {
		if !domains[i].Privacy {
			noPrivacyDomains = append(noPrivacyDomains, domains[i])
		}
	}
	return noPrivacyDomains
}
