package httpapi

import (
	"net/http"
	"strconv"

	httpdto "mail-sync-service/internal/controller/http/v1/dto"
	"mail-sync-service/internal/service"

	"github.com/labstack/echo/v4"
)

type SearchHandler struct {
	searchService service.Search
}

func newSearchRoutes(g *echo.Group, searchService service.Search) {
	h := &SearchHandler{searchService: searchService}
	g.GET("", h.Search)
}

func (h *SearchHandler) Search(c echo.Context) error {
	ctx := c.Request().Context()
	q := c.QueryParam("q")
	mailboxID := c.QueryParam("mailbox_id")
	folder := c.QueryParam("folder")
	from := c.QueryParam("from")
	to := c.QueryParam("to")
	dateFrom := c.QueryParam("date_from")
	dateTo := c.QueryParam("date_to")
	pageStr := c.QueryParam("page")
	sizeStr := c.QueryParam("size")
	hasAttachStr := c.QueryParam("has_attachments")
	seenStr := c.QueryParam("seen")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var hasAttach *bool
	if hasAttachStr != "" {
		b, _ := strconv.ParseBool(hasAttachStr)
		hasAttach = &b
	}
	var seen *bool
	if seenStr != "" {
		b, _ := strconv.ParseBool(seenStr)
		seen = &b
	}

	input := service.SearchInput{
		Query:     q,
		MailboxID: mailboxID,
		Folder:    folder,
		From:      from,
		To:        to,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		HasAttach: hasAttach,
		Seen:      seen,
		Page:      page,
		Size:      size,
	}
	items, total, err := h.searchService.Search(ctx, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpdto.ErrorOutput{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	})
}
