package services

import (
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func checkUserDomainPermission(uId interface{}, dId interface{}, target models.UserDomainRole) bool {
	var ud models.UserDomain
	if err := db.DB.Where("user_id = ? AND domain_id = ?", uId, dId).First(&ud).Error; err != nil {
		return false
	}
	return ud.Role >= target
}

// @Summary Create UserDomain Relation
// @Description Create UserDomain Relation
// @Description user must have manager permission to domain or be admin
// @Description user cant create permission higher than himself
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Param userRole body mw.DomainUser true "userRole"
// @Produce json
// @Success 200 {object} mw.Domain{data=models.UserDomain}
// @Failure 400 {object} mw.Domain{data=int}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id}/user [post]
func UserDomainCreate(c *fiber.Ctx) error {
	// get domainId restful api
	qId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: "invalid domain id",
			Data:   nil,
		})
	}

	// get userRole from body
	var userRole mw.DomainUser
	if err := c.BodyParser(&userRole); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: "invalid request body",
			Data:   nil,
		})
	}

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user admin
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)

	// check if user can access domain
	if !(flag || checkUserDomainPermission(uId, strconv.Itoa(qId),
		utils.IfThen(userRole.Role >= models.Manager, models.Owner, models.Manager))) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// add user to domain
	if err := db.DB.Create(&models.UserDomain{
		UserId:   uint(userRole.UserId),
		DomainId: uint(qId),
		Role:     userRole.Role,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   userRole,
	})
}

// @Summary Delete UserDomain Relation
// @Description Delete UserDomain Relation **(no update, just delete and create)**
// @Description user must have manager permission to domain or be admin
// @Description user cant delete permission higher than himself
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Param uid path string true "user id"
// @Produce json
// @Success 200 {object} mw.Domain{data=int}
// @Failure 400 {object} mw.Domain{data=int}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id}/user/{uid} [delete]
func UserDomainDelete(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get user from params
	quId := c.Params("uid")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user admin
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.Manager)) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// delete user from domain
	if checkUserDomainPermission(uId, qId, models.Owner) {
		if err := db.DB.Where("user_id = ? AND domain_id = ?", quId, qId).Delete(&models.UserDomain{}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   nil,
			})
		}
	} else {
		userDomain := models.UserDomain{}
		if err := db.DB.Where("user_id = ? AND domain_id = ? AND role < ?", quId, qId, models.Owner).First(&userDomain).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
				Status: fiber.StatusForbidden,
				Errors: "permission denied",
				Data:   qId,
			})
		}
		if err := db.DB.Delete(&userDomain).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   nil,
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   nil,
	})
}

// no update, just delete and create

// @Summary List UserDomain Relation
// @Description List UserDomain Relation
// @Description user must have manager permission to domain or be admin
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Produce json
// @Success 200 {object} mw.Domain{data=[]mw.DomainUserDetail}
// @Failure 400 {object} mw.Domain{data=int}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id}/user [get]
func UserDomainList(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

	// check if user admin
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.Manager)) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// get user list
	var users []mw.DomainUserDetail
	if err := db.DB.Table("user_domains").
		Select([]string{
			"`user_domains`.`user_id` as `user_id`",
			"`users`.`name` as `username`",
			"`users`.`email` as `email`",
			"`user_domains`.`role` as `role`",
			"`user_domains`.`domain_id` as `domain_id`",
			"`domains`.`name` as `domain_name`",
		}).
		Joins("left join users on users.id = user_domains.user_id").
		Joins("left join domains on domains.id = user_domains.domain_id").
		Where("user_domains.domain_id = ?", qId).
		Scan(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   users,
	})
}
