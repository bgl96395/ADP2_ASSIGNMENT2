package http

import (
	"errors"
	"net/http"

	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/model"
	"github.com/bgl96395/ADP2_ASSIGNMENT2/appointment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Appointment_handler struct {
	usecase *usecase.Appointment_usecase
}

func New_appointment_handler(usecase *usecase.Appointment_usecase) *Appointment_handler {
	return &Appointment_handler{usecase: usecase}
}

func (handler *Appointment_handler) Register_routes(register *gin.Engine) {
	register.POST("/appointments", handler.Create)
	register.GET("/appointments", handler.Get_all)
	register.GET("/appointments/:id", handler.Get_by_ID)
	register.PATCH("/appointments/:id/status", handler.Update_Status)
}

func (handler *Appointment_handler) Create(context *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		DoctorID    int    `json:"doctor_id"`
	}
	err := context.ShouldBindJSON(&req)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	appointment, err := handler.usecase.Create_appointment(req.Title, req.Description, req.DoctorID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, usecase.E_doctor_service_unavailable) {
			status = http.StatusServiceUnavailable
		}
		context.JSON(status, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, appointment)
}

func (handler *Appointment_handler) Get_by_ID(context *gin.Context) {
	id := context.Param("id")
	appointment, err := handler.usecase.Get_appointment(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, appointment)
}

func (handler *Appointment_handler) Get_all(context *gin.Context) {
	list, err := handler.usecase.List_appoinments()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, list)
}

func (handler *Appointment_handler) Update_Status(context *gin.Context) {
	id := context.Param("id")
	var req struct {
		Status model.Status `json:"status"`
	}
	err := context.ShouldBindJSON(&req)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	appointment, err := handler.usecase.Update_status(id, req.Status)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, appointment)
}
