package acceptor

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PaginationParams struct {
	Page    int64 `form:"page,default=1"`
	PerPage int64 `form:"per_page,default=20"`
}

func bindRequestParams(c *fiber.Ctx, p *PaginationParams) error {
	pageParam := c.Params("page", "1")
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		return err
	}

	perPageParam := c.Params("per_page", "20")
	perPage, err := strconv.Atoi(perPageParam)
	if err != nil {
		return err
	}

	p.Page = int64(page)
	p.PerPage = int64(perPage)
	return nil
}
