package services

import (
	// md "domain0/modules/dns"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/modules"
)

// @Summary List Domain Dns
// @Description List Domain Dns **AliDNS as Example, read modules for others**
// @Description user must have read permission to domain or be admin
// @Tags domain
// @Accept json
// @Param id path string true "domain id"
// @Produce json
// @Success 200 {object} mw.Domain{data=[]md.CloudflareDNSList}
// @Success 200 {object} mw.Domain{data=[]md.TencentDNSList}
// @Success 200 {object} mw.Domain{data=[]md.AliDNSList}
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
	isAdmin := c.Locals("role").(models.UserRole) >= models.Admin
	permission := checkUserDomainPermission(uId, qId, models.ReadOnly)
	if !(isAdmin || permission) {
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

	// admin have no access to privacy domain
	if !permission && isAdmin && domain.Privacy {
		return c.Status(fiber.StatusForbidden).JSON(mw.Domain{
			Status: fiber.StatusForbidden,
			Errors: "permission denied, privacy domain",
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
	flag := c.Locals("role").(models.UserRole) >= models.Admin
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
	// get dns record
	dnsObj := modules.DnsObjGen(&domain)
	if err := dnsObj.Get(dnsId); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "dns record not found",
			Data:   qId,
		})
	}

	if err := dnsObj.Delete(); err != nil {
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
// @Param dns body md.AliDNS true "dns info"
// @Produce json
// @Success 200 {object} mw.Domain{data=md.AliDNS}
// @Success 200 {object} mw.Domain{data=md.TencentDNS}
// @Success 200 {object} mw.Domain{data=md.CloudflareDNS}
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
	flag := c.Locals("role").(models.UserRole) >= models.Admin
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
		dnsObjson, err := json.Marshal(modules.DnsChangeStruct{
			Dns:    dnsObj,
			Domain: domain,
		})
		if err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: err.Error(),
				Data:   qId,
			})
		}
		nqId, _ := strconv.Atoi(qId)
		dc := models.DomainChange{
			DomainId:     uint(nqId),
			UserId:       uId,
			ActionType:   models.Submit,
			ActionStatus: models.Reviewing,
			Reason:       fmt.Sprintf("%d want to create dns record for domain %s:", uId, domain.Name),
			Operation:    string(dnsObjson),
		}
		if err := db.DB.Create(&dc).Error; err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: err.Error(),
				Data:   qId,
			})
		}
		return c.Status(fiber.StatusAlreadyReported).JSON(mw.Domain{
			Status: fiber.StatusAlreadyReported,
			Data:   "ICP domain need owner permission, please wait for approval",
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
// @Param dns body md.AliDNS true "dns info"
// @Produce json
// @Success 200 {object} mw.Domain{data=md.AliDNS}
// @Success 200 {object} mw.Domain{data=md.TencentDNS}
// @Success 200 {object} mw.Domain{data=md.CloudflareDNS}
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
	flag := c.Locals("role").(models.UserRole) >= models.Admin
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
	dnsObj := modules.DnsObjGen(&domain)
	if err := dnsObj.Get(dnsId); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusNotFound).JSON(mw.Domain{
			Status: fiber.StatusNotFound,
			Errors: "dns record not found",
			Data:   qId,
		})
	}

	// update dns record
	if err := c.BodyParser(dnsObj); err != nil {
		logrus.Error(err)
		return c.Status(fiber.StatusBadRequest).JSON(mw.Domain{
			Status: fiber.StatusBadRequest,
			Errors: err.Error(),
			Data:   qId,
		})
	}

	if domain.ICPReg > 0 && !checkUserDomainPermission(uId, qId, models.Owner) {
		dnsObjson, err := json.Marshal(modules.DnsChangeStruct{
			Dns:    dnsObj,
			Domain: domain,
		})
		if err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: err.Error(),
				Data:   qId,
			})
		}
		nqId, _ := strconv.Atoi(qId)
		dc := models.DomainChange{
			DomainId:     uint(nqId),
			UserId:       uId,
			ActionType:   models.EditDNS,
			ActionStatus: models.Reviewing,
			Reason:       fmt.Sprintf("%d want to update dns record for domain %s:", uId, domain.Name),
			Operation:    string(dnsObjson),
		}
		if err := db.DB.Create(&dc).Error; err != nil {
			logrus.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.Domain{
				Status: fiber.StatusInternalServerError,
				Errors: err.Error(),
				Data:   qId,
			})
		}
		return c.Status(fiber.StatusAlreadyReported).JSON(mw.Domain{
			Status: fiber.StatusAlreadyReported,
			Data:   "ICP domain need owner permission, please wait for approval",
		})
	} else {
		if err := dnsObj.Update(); err != nil {
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
