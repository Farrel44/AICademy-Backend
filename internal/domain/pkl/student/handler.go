package pkl

import (
	"fmt"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type StudentPklHandler struct {
	service *StudentPklService
}

func NewStudentPklHandler(service *StudentPklService) *StudentPklHandler {
	return &StudentPklHandler{
		service: service,
	}
}

func (h *StudentPklHandler) ApplyPklPosition(c *fiber.Ctx) error {
	var req ApplyInternshipRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	fmt.Printf("internship id on handler : %s", &req.InternshipID)

	result, err := h.service.ApplyStudentInternshipPosition(c, req.InternshipID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Data siswa berhasil diambil", result)
}
