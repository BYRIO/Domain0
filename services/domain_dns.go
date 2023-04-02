package services

import (
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/modules"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// @Summary List Domain Dns
// @Description List Domain Dns **AliDNS as Example, read modules for others**
// @Description user must have read permission to domain or be admin
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Produce json
// @Success 200 {object} mw.Domain{data=[]modules.CloudflareDNSList}
// @Success 200 {object} mw.Domain{data=[]modules.TencentDNSList}
// @Success 200 {object} mw.Domain{data=[]modules.AliDNSList}
// @Failure 400 {object} mw.Domain{data=int}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id}/dns [get]
func DomainDnsList(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

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

// @Summary Delete Domain Dns
// @Description Delete Domain Dns
// @Description user must have readwrite permission to domain or be admin
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Param dnsId path string true "dns id"
// @Produce json
// @Success 200 {object} mw.Domain{data=int}
// @Failure 400 {object} mw.Domain{data=int}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Failure 500 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id}/dns/{dnsId} [delete]
func DomainDnsDelete(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

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

	dnsObj := []interface{}{}
	if err := dnsObjList.MultipleSelectWithIds([]string{dnsId}, &dnsObj); err != nil || len(dnsObj) != 1 {
		logrus.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	if err := dnsObj[0].(modules.DnsObj).Delete(); err != nil {
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

// @Summary Create Domain Dns
// @Description Create Domain Dns **AliDNS as Example, read modules for others**
// @Description user must have readwrite permission to domain or be admin
// @Description for now only owner can edit domain which ICP_reg is true
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Param dns body modules.AliDNS true "dns info"
// @Produce json
// @Success 200 {object} mw.Domain{data=modules.AliDNS}
// @Failure 400 {object} mw.Domain{data=int}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Failure 500 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id}/dns [post]
func DomainDnsCreate(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

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

	if domain.ICPReg > 0 && !checkUserDomainPermission(uId, qId, models.Owner) {
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

// @Summary Update Domain Dns
// @Description Update Domain Dns **AliDNS as Example, read modules for others**
// @Description user must have readwrite permission to domain or be admin
// @Description for now only owner can edit domain which ICP_reg is true
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Param dnsId path string true "dns id"
// @Param dns body modules.AliDNS true "dns info"
// @Produce json
// @Success 200 {object} mw.Domain{data=modules.AliDNS}
// @Failure 400 {object} mw.Domain{data=int}
// @Failure 403 {object} mw.Domain{data=int}
// @Failure 404 {object} mw.Domain{data=int}
// @Failure 500 {object} mw.Domain{data=int}
// @Router /api/v1/domain/{id}/dns/{dnsId} [put]
func DomainDnsUpdate(c *fiber.Ctx) error {
	// get domainId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(uint)

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

	dnsObj := []interface{}{}
	if err := dnsObjList.MultipleSelectWithIds([]string{dnsId}, &dnsObj); err != nil || len(dnsObj) != 1 {
		logrus.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
			Status: fiber.StatusInternalServerError,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	// update dns record
	if err := c.BodyParser(dnsObj[0]); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	if domain.ICPReg > 0 && !checkUserDomainPermission(uId, qId, models.Owner) {
		// todo: notify
		return c.Status(fiber.StatusNotImplemented).JSON(mw.Domain{
			Status: fiber.StatusNotImplemented,
			Errors: "ICP domain need owner permission",
			Data:   qId,
		})
	} else {
		if err := dnsObj[0].(modules.DnsObj).Update(); err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: err.Error(),
				Data:   qId,
			})
		}

		logrus.Info("User: ", uId, " update dns record: ", dnsObj[0], " for domain: ", qId)
		return c.JSON(mw.Domain{
			Status: fiber.StatusOK,
			Data:   dnsObj[0],
		})
	}
}
