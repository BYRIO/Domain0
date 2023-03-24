package services

import (
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func UserInfoGet(c *fiber.Ctx) error {
	// get userId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if query user is the same as jwt user, or if jwt user is admin
	if qId != uId && c.Locals("role").(models.UserRole) < models.Admin {
		return c.Status(fiber.StatusForbidden).JSON(mw.User{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   uId,
		})
	}

	// query user info
	var user models.User
	if err := db.DB.Where("id = ?", qId).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(mw.User{
			Status: fiber.StatusNotFound,
			Errors: "user not found",
			Data:   uId,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.User{
		Status: fiber.StatusOK,
		Errors: "",
		Data:   user,
	})
}

func UserInfoUpdate(c *fiber.Ctx) error {
	// get userId restful api
	qId := c.Params("id")

	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if query user is the same as jwt user, or if jwt user is admin
	if qId != uId && c.Locals("role").(models.UserRole) < models.Admin {
		return c.Status(fiber.StatusForbidden).JSON(mw.User{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   uId,
		})
	}

	// query user info
	var user models.User
	if err := db.DB.Where("id = ?", qId).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(mw.User{
			Status: fiber.StatusNotFound,
			Errors: "user not found",
			Data:   qId,
		})
	}

	// update user info
	var updateInfo mw.UserInfoUpdate
	if err := c.BodyParser(&updateInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mw.User{
			Status: fiber.StatusBadRequest,
			Errors: "bad request",
			Data:   uId,
		})
	}

	// update user info
	if c.Locals("role").(models.UserRole) < models.Admin {
		if updateInfo.Name != nil || updateInfo.StuId != nil {
			return c.Status(fiber.StatusForbidden).JSON(mw.User{
				Status: fiber.ErrNotImplemented,
				Errors: "permission denied for now",
			})
		} // only admin can update name and stuid, in the future, we may allow user to update name and stuid with check
		if updateInfo.Role != nil {
			logrus.Warnf("user %s try to update role of user %s", uId, qId)
			return c.Status(fiber.StatusForbidden).JSON(mw.User{
				Status: fiber.ErrForbidden,
				Errors: "permission denied, you've been reported",
			})
		} // only admin can update role
	} else {
		if updateInfo.Role != nil && c.Locals("role").(models.UserRole) <= updateInfo.Role.(models.UserRole) {
			logrus.Warnf("user %s try to overstep update role of user %s", uId, qId)
			return c.Status(fiber.StatusForbidden).JSON(mw.User{
				Status: fiber.ErrForbidden,
				Errors: "permission denied, you've been reported",
			})
		} // admin can't update role to the same or higher than himself
	}

	user.Email = utils.IfThen(updateInfo.Email != nil, updateInfo.Email.(string), user.Email)
	user.Name = utils.IfThen(updateInfo.Name != nil, updateInfo.Name.(string), user.Name)
	user.StuId = utils.IfThen(updateInfo.StuId != nil, updateInfo.StuId.(string), user.StuId)
	user.Role = utils.IfThen(updateInfo.Role != nil, updateInfo.Role.(models.UserRole), user.Role)
	if updateInfo.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateInfo.Password.(string)), bcrypt.MaxCost)
		if err != nil {
			logrus.Errorf("bcrypt password error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(mw.User{
				Status: fiber.StatusInternalServerError,
				Errors: "internal server error",
				Data:   uId,
			})
		}
		user.Password = string(hashedPassword)
	}

	if err := db.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mw.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   uId,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.User{
		Status: fiber.StatusOK,
		Errors: "",
		Data:   user,
	})
}
