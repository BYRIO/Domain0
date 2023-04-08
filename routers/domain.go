package routers

import (
	"domain0/services"

	"github.com/gofiber/fiber/v2"
)

func SetupDomainRouter(r fiber.Router) {
	domain := r.Group("/domain")
	domain.Get("/", services.DomainList)
	domain.Get("/:id", services.DomainGet)
	domain.Put("/:id", services.DomainUpdate)
	domain.Delete("/:id", services.DomainDelete)
	domain.Post("/", services.DomainCreate)

	SetupUserDomainRouter(domain)
	SetupDomainDnsRouter(domain)
	SetupDomainChangeRouter(domain)
}

func SetupUserDomainRouter(r fiber.Router) {
	r.Get(":id/user", services.UserDomainList)
	r.Post(":id/user", services.UserDomainCreate)
	r.Delete(":id/user/:uid", services.UserDomainDelete)
}

func SetupDomainDnsRouter(r fiber.Router) {
	r.Get(":id/dns", services.DomainDnsList)
	r.Post(":id/dns", services.DomainDnsCreate)
	r.Put(":id/dns/:dnsId", services.DomainDnsUpdate)
	r.Delete(":id/dns/:dnsId", services.DomainDnsDelete)
}

func SetupDomainChangeRouter(r fiber.Router) {
	r.Get("/change/myapply", services.DomainChangeListMyApply)
	r.Get("/change/myapprove", services.DomainChangeListMyApprove)
	r.Get("/change/:id", services.DomainChangeCheck)
}
