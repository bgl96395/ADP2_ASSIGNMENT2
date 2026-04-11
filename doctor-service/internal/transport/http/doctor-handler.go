package http

import (
	"net/http"

	"doctor-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Doctor_handler struct {
	usecase *usecase.Doctor_usecase
}

func New_doctor_handler(usecase *usecase.Doctor_usecase) *Doctor_handler {
	return &Doctor_handler{usecase: usecase}
}

func (handler *Doctor_handler) RegisterRoutes(repository *gin.Engine) {
	repository.POST("/doctors", handler.Create)
	repository.GET("/doctors/:id", handler.Get_by_ID)
	repository.GET("/doctors", handler.List)
}

type create_doctor_request struct {
	FullName       string `json:"full_name"`
	Specialization string `json:"specialization"`
	Email          string `json:"email"`
}

func (handler *Doctor_handler) Create(context *gin.Context) {
	var req create_doctor_request
	err := context.ShouldBindJSON(&req)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	doctor, err := handler.usecase.Create_doctor(req.FullName, req.Specialization, req.Email)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, doctor)
}

func (handler *Doctor_handler) Get_by_ID(context *gin.Context) {
	id := context.Param("id")
	doctor, err := handler.usecase.Get_doctor(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, doctor)
}

func (handler *Doctor_handler) List(context *gin.Context) {
	list, err := handler.usecase.List_doctors()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, list)
}
