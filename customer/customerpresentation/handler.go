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

func (h *CustomerHandler) ContactPage(c *gin.Context) {
	c.HTML(http.StatusOK, "contact.html", nil)
}

func (h *CustomerHandler) ShippingInfoPage(c *gin.Context) {
	renderInfoPage(c,
		"Shipping Info",
		"Every order is reviewed with you before production, and dispatch timing depends on your final design approval.",
		[]gin.H{
			{"heading": "Production Timeline", "body": "Wedding cards are produced after the final proof is approved. Lead times vary by design complexity, quantity, and finishing options such as foil and inserts."},
			{"heading": "Delivery Coverage", "body": "We coordinate deliveries across Pakistan and can discuss international shipping separately. If you need a firm delivery window, contact us before placing the order."},
			{"heading": "Support", "body": "For urgent delivery questions, message us on WhatsApp at +92 303 9238299 before checkout so we can confirm the schedule."},
		},
	)
}

func (h *CustomerHandler) ReturnsExchangesPage(c *gin.Context) {
	renderInfoPage(c,
		"Returns & Exchanges",
		"Because most orders are personalized, returns are handled case by case after we review the production status and the issue reported.",
		[]gin.H{
			{"heading": "Custom Orders", "body": "Personalized and approved print jobs usually cannot be returned once production has started, but we will investigate any quality issue or fulfillment error promptly."},
			{"heading": "Before Production", "body": "If you need to change or cancel an order before approval and printing, contact us immediately so we can confirm what is still possible."},
			{"heading": "Resolution", "body": "If something arrives damaged or incorrect, share your order number and photos with our team and we will review the next step as quickly as possible."},
		},
	)
}

func (h *CustomerHandler) MyAccountPage(c *gin.Context) {
	renderInfoPage(c,
		"My Account",
		"Customer accounts are not enabled yet. Orders are currently managed directly through your order number and our support channels.",
		[]gin.H{
			{"heading": "Track an Order", "body": "Use your order number on the track-order page to view the latest order status and personalization details."},
			{"heading": "Need Help", "body": "If you need any change, confirmation, or production update, contact us through WhatsApp or email and include your order number."},
		},
	)
}

func (h *CustomerHandler) TermsOfUsePage(c *gin.Context) {
	renderInfoPage(c,
		"Terms of Use",
		"By placing an order with Write&InviteCo, you confirm that the submitted personalization details are accurate and approved for production.",
		[]gin.H{
			{"heading": "Design Approval", "body": "Production starts only after the final design details are confirmed. Review names, dates, venues, and contact details carefully before approval."},
			{"heading": "Pricing", "body": "Final pricing depends on the selected product, quantity, foil option, and requested inserts. Any custom change outside the standard flow may require a revised quote."},
			{"heading": "Communication", "body": "We may contact you by phone, email, or WhatsApp to confirm the order, clarify details, and share status updates."},
		},
	)
}

func (h *CustomerHandler) PrivacyPolicyPage(c *gin.Context) {
	renderInfoPage(c,
		"Privacy Policy",
		"We collect the customer and event details required to personalize invitations, communicate about orders, and provide status updates.",
		[]gin.H{
			{"heading": "What We Store", "body": "Your name, contact details, address, and event information are stored so your order can be produced and tracked accurately."},
			{"heading": "How We Use It", "body": "Your data is used for order fulfillment, communication, and confirmation updates. We do not expose submitted details publicly."},
			{"heading": "Questions", "body": "If you need a correction to submitted order details, contact us directly with your order number so we can help."},
		},
	)
}

func renderInfoPage(c *gin.Context, title string, intro string, sections []gin.H) {
	c.HTML(http.StatusOK, "info.html", gin.H{
		"title":    title,
		"intro":    intro,
		"sections": sections,
	})
}
