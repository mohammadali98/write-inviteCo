package customerpresentation

import (
	"net/http"

	"writeandinviteco/inviteandco/customer/customerdomain"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	repo customerdomain.CustomerRepo
}

func NewCustomerHandler(repo customerdomain.CustomerRepo) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

func (h *CustomerHandler) AboutPage(c *gin.Context) {
	c.HTML(http.StatusOK, "about.html", nil)
}
