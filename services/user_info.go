package services

import (
	"database/sql"
	db "domain0/database"
	"domain0/models"
	mw "domain0/models/web"
	"domain0/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Get user info
// @Description Get user info by id
// @Description if id is not the same as jwt sub, jwt role must be admin
// @Tags user
// @Param id path string true "user id"
// @Produce json
// @Success 200 {object} mw.User{data=models.User}
// @Failure 403 {object} mw.User{data=int}
// @Failure 404 {object} mw.User{data=int}
// @Router /api/v1/user/{id} [get]
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
		Data:   user,
	})
}

// @Summary Update user info
// @Description Update user info by id
// @Description Only admin can update other user info whoes role is lower than his.
// @Tags user
// @Param id path string true "user id"
// @Accept json
// @Produce json
// @Param user body models.User true "user info"
// @Success 200 {object} mw.User{data=models.User}
// @Failure 400 {object} mw.User{data=int}
// @Failure 403 {object} mw.User{data=int}
// @Failure 404 {object} mw.User{data=int}
// @Failure 500 {object} mw.User{data=int}
// @Router /api/v1/user/{id} [put]
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
				Status: fiber.StatusNotImplemented,
				Errors: "permission denied for now",
			})
		} // only admin can update name and stuid, in the future, we may allow user to update name and stuid with check
		if updateInfo.Role != nil {
			logrus.Warnf("user %s try to update role of user %s", uId, qId)
			return c.Status(fiber.StatusForbidden).JSON(mw.User{
				Status: fiber.StatusForbidden,
				Errors: "permission denied, you've been reported",
			})
		} // only admin can update role
	} else {
		if updateInfo.Role != nil && c.Locals("role").(models.UserRole) <= *updateInfo.Role {
			logrus.Warnf("user %s try to overstep update role of user %s", uId, qId)
			return c.Status(fiber.StatusForbidden).JSON(mw.User{
				Status: fiber.StatusForbidden,
				Errors: "permission denied, you've been reported",
			})
		} // admin can't update role to the same or higher than himself
	}

	user.Email = utils.IfThen(updateInfo.Email != nil, *updateInfo.Email, user.Email)
	user.Name = utils.IfThen(updateInfo.Name != nil, *updateInfo.Name, user.Name)
	user.StuId = utils.IfThen(updateInfo.StuId != nil, sql.NullString{
		String: *updateInfo.StuId,
		Valid:  true,
	}, user.StuId)
	user.Role = utils.IfThen(updateInfo.Role != nil, *updateInfo.Role, user.Role)
	if updateInfo.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updateInfo.Password), bcrypt.MaxCost)
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

// @Summary Delete user
// @Description Delete user by id
// @Description Only admin can delete user, and the user must has none role.
// @Tags user
// @Param id path string true "user id"
// @Accept json
// @Produce json
// @Success 200 {object} mw.User{data=int}
// @Failure 400 {object} mw.User{data=int}
// @Failure 403 {object} mw.User{data=int}
// @Failure 404 {object} mw.User{data=int}
// @Failure 500 {object} mw.User{data=int}
// @Router /api/v1/user/{id} [delete]
func UserInfoDelete(c *fiber.Ctx) error {
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

	// if user has role, can't delete
	if user.Role != models.Normal {
		return c.Status(fiber.StatusForbidden).JSON(mw.User{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   qId,
		})
	}

	// delete user
	if err := db.DB.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mw.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   uId,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.User{
		Status: fiber.StatusOK,
		Errors: "",
		Data:   uId,
	})
}

// @Summary Get user list
// @Description Get user list
// @Description Only admin can get user list.
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} mw.User{data=[]models.User}
// @Failure 400 {object} mw.User{data=int}
// @Failure 403 {object} mw.User{data=int}
// @Failure 404 {object} mw.User{data=int}
// @Failure 500 {object} mw.User{data=int}
// @Router /api/v1/user [get]
func UserList(c *fiber.Ctx) error {
	// get query user info from jwt sub
	uId := c.Locals("sub").(string)

	// check if jwt user is admin
	if c.Locals("role").(models.UserRole) < models.Admin {
		return c.Status(fiber.StatusForbidden).JSON(mw.User{
			Status: fiber.StatusForbidden,
			Errors: "permission denied",
			Data:   uId,
		})
	}

	// query user list
	var users []models.User
	if err := db.DB.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(mw.User{
			Status: fiber.StatusInternalServerError,
			Errors: "internal server error",
			Data:   uId,
		})
	}

	return c.Status(fiber.StatusOK).JSON(mw.User{
		Status: fiber.StatusOK,
		Errors: "",
		Data:   users,
	})
}
