package services

import (
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/modules"

	"github.com/gofiber/fiber/v2"
)

func DomainDnsList(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if user role level
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.ReadOnly)) {
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

	// get domain dns list
	dnsList := modules.DnsListObjGen(&domain)
	if err := dnsList.GetDNSList(&domain); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	return c.JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   dnsList,
	})
}
