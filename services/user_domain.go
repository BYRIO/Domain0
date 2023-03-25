package services

import (
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func checkUserDomainPermission(uId string, dId string, target models.UserDomainRole) bool {
	var ud models.UserDomain
	if err := db.DB.Where("user_id = ? AND domain_id = ?", uId, dId).First(&ud).Error; err != nil {
		return false
	}
	return ud.Role >= target
}

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
	uId := c.Locals("sub").(string)

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

func UserDomainDelete(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get user from params
	quId := c.Params("uid")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if user admin
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.Owner)) {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// delete user from domain
	if err := db.DB.Where("user_id = ? AND domain_id = ?", quId, qId).Delete(&models.UserDomain{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   nil,
	})
}

// no update, just delete and create

func UserDomainList(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

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
		Select("user_domains.user_id, users.username, users.email, user_domains.role, user_domains.domain_id, domains.name").
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
