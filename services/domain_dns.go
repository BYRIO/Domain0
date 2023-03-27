package services

import (
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/modules"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func DomainDnsList(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if user role level
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.ReadOnly)) {
		logrus.Info("User: ", uId, " try to access domain: ", qId, " without permission")
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
		logrus.Error(err)
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

func DomainDnsDelete(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if user role level
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.ReadWrite)) {
		logrus.Info("User: ", uId, " try to access domain: ", qId, " without permission")
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// get domain info
	var domain models.Domain
	if err := db.DB.Where("id = ?", qId).First(&domain).Error; err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "domain not found",
			Data:   qId,
		})
	}

	// get dns record id
	dnsId := c.Params("dnsId")

	// delete dns record
	dnsObjList := modules.DnsListObjGen(&domain)
	if err := dnsObjList.GetDNSList(&domain); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	dnsObj := interface{}(modules.DnsObjGen(&domain))
	if err := dnsObjList.MultipleSelectWithIds([]string{dnsId}, &dnsObj); err != nil || len(dnsObj.([]interface{})) != 1 {
		logrus.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	if err := dnsObj.([]modules.DnsObj)[0].Delete(); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	return c.JSON(mw.Domain{
		Status: fiber.StatusOK,
		Data:   qId,
	})
}

func DomainDnsCreate(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if user role level
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.ReadWrite)) {
		logrus.Info("User: ", uId, " try to access domain: ", qId, " without permission")
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// get domain info
	var domain models.Domain
	if err := db.DB.Where("id = ?", qId).First(&domain).Error; err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "domain not found",
			Data:   qId,
		})
	}

	// generate dns record
	dnsObj := modules.DnsObjGen(&domain)
	if err := c.BodyParser(dnsObj); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	if domain.ICPReg && !checkUserDomainPermission(uId, qId, models.Owner) {
		// todo: notify
		return c.Status(fiber.StatusNotImplemented).JSON(mw.Domain{
			Status: fiber.StatusNotImplemented,
			Errors: "ICP domain need owner permission",
			Data:   qId,
		})

	} else {
		// create dns record
		if err := dnsObj.Create(); err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: err.Error(),
				Data:   qId,
			})
		}

		logrus.Info("User: ", uId, " create dns record: ", dnsObj, " for domain: ", qId)
		return c.JSON(mw.Domain{
			Status: fiber.StatusCreated,
			Data:   dnsObj,
		})
	}
}

func DomainDnsUpdate(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if user role level
	flag := (c.Locals("role").(models.UserRole) >= models.Admin)
	if !(flag || checkUserDomainPermission(uId, qId, models.ReadWrite)) {
		logrus.Info("User: ", uId, " try to access domain: ", qId, " without permission")
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// get domain info
	var domain models.Domain
	if err := db.DB.Where("id = ?", qId).First(&domain).Error; err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "domain not found",
			Data:   qId,
		})
	}

	// get dns record id
	dnsId := c.Params("dnsId")

	// get dns record
	dnsObjList := modules.DnsListObjGen(&domain)
	if err := dnsObjList.GetDNSList(&domain); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	dnsObj := interface{}(modules.DnsObjGen(&domain))
	if err := dnsObjList.MultipleSelectWithIds([]string{dnsId}, &dnsObj); err != nil || len(dnsObj.([]interface{})) != 1 {
		logrus.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	// update dns record
	if err := c.BodyParser(dnsObj.([]modules.DnsObj)[0]); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	if domain.ICPReg && !checkUserDomainPermission(uId, qId, models.Owner) {
		// todo: notify
		return c.Status(fiber.StatusNotImplemented).JSON(mw.Domain{
			Status: fiber.StatusNotImplemented,
			Errors: "ICP domain need owner permission",
			Data:   qId,
		})
	} else {
		if err := dnsObj.([]modules.DnsObj)[0].Update(); err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: err.Error(),
				Data:   qId,
			})
		}

		logrus.Info("User: ", uId, " update dns record: ", dnsObj, " for domain: ", qId)
		return c.JSON(mw.Domain{
			Status: fiber.StatusOK,
			Data:   dnsObj,
		})
	}
}
